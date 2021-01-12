package environment

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/stellentus/cartpoles/lib/util/convformat"
	transModel "github.com/stellentus/cartpoles/lib/util/kdtree"
	"github.com/stellentus/cartpoles/lib/util/random"
	tpo "github.com/stellentus/cartpoles/lib/util/type-opr"

	"github.com/stellentus/cartpoles/lib/logger"
	"github.com/stellentus/cartpoles/lib/rlglue"
)

type knnSettings struct {
	DataLog      string  `json:"datalog"`
	Seed         int64   `json:"seed"`
	Neighbor_num int     `json:"neighbor-num"`
	EnsembleSeed int     `json:"ensemble-seed"`
	DropPerc     float64 `json:"drop-percent"`
}

type KnnModelEnv struct {
	logger.Debug
	knnSettings
	state           rlglue.State
	rng             *rand.Rand
	offlineData     [][]float64
	offlineStarts   []int
	offlineModel    transModel.TransTrees
	stateDim        int
	NumberOfActions int
	stateRange      []float64
	neighbor_prob   []float64
}

func init() {
	Add("knnModel", NewKnnModelEnv)
}

func NewKnnModelEnv(logger logger.Debug) (rlglue.Environment, error) {
	return &KnnModelEnv{Debug: logger}, nil
}

func (env *KnnModelEnv) SettingFromLog(paramLog string) {
	// Get state dimension
	txt, err := os.Open(paramLog)
	if err != nil {
		panic(err)
	}
	defer txt.Close()
	scanner := bufio.NewScanner(txt)
	var line string
	var spl []string
	var stateDim int
	var numAction int
	var stateRange []float64
	for scanner.Scan() {
		line = scanner.Text()
		spl = strings.Split(line, "=")
		if spl[0] == "stateDimension" { //stateDimension
			stateDim, _ = strconv.Atoi(spl[1])
		} else if spl[0] == "numberOfActions" {
			numAction, _ = strconv.Atoi(spl[1])
		} else if spl[0] == "stateRange" {
			stateRange = convformat.ListStr2Float(spl[1], ",")
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	env.stateDim = stateDim
	env.NumberOfActions = numAction
	env.stateRange = stateRange
}

// Initialize configures the environment with the provided parameters and resets any internal state.
func (env *KnnModelEnv) Initialize(run uint, attr rlglue.Attributes) error {
	err := json.Unmarshal(attr, &env.knnSettings)
	if err != nil {
		err = errors.New("environment.knnModel settings error: " + err.Error())
		env.Message("err", err)
		return err
	}
	env.Seed += int64(run)
	env.rng = rand.New(rand.NewSource(env.Seed)) // Create a new rand source for reproducibility
	env.state = make(rlglue.State, 4)

	env.Message("environment.knnModel settings", fmt.Sprintf("%+v", env.knnSettings))

	folder := env.knnSettings.DataLog
	traceLog := folder + "/traces-" + strconv.Itoa(int(run)) + ".csv"
	paramLog := folder + "/log_json.txt"
	env.SettingFromLog(paramLog)

	// Get offline data
	csvFile, err := os.Open(traceLog)
	if err != nil {
		log.Fatal(err)
	}
	allTransStr, err := csv.NewReader(csvFile).ReadAll()
	csvFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	allTransTemp := make([][]float64, len(allTransStr)-1)
	for i := 1; i < len(allTransStr); i++ { // remove first str (title of column)
		trans := allTransStr[i]
		row := make([]float64, env.stateDim*2+3)
		for j, num := range trans {
			if j == 0 { // next state
				num = num[1 : len(num)-1] // remove square brackets
				copy(row[env.stateDim+1:env.stateDim*2+1], convformat.ListStr2Float(num, " "))
			} else if j == 1 { // current state
				num = num[1 : len(num)-1]
				copy(row[:env.stateDim], convformat.ListStr2Float(num, " "))
			} else if j == 2 { // action
				row[env.stateDim], _ = strconv.ParseFloat(num, 64)
			} else if j == 3 { //reward
				row[env.stateDim*2+1], _ = strconv.ParseFloat(num, 64)
				if row[env.stateDim*2+1] == -1 { // termination
					row[env.stateDim*2+2] = 1
				} else {
					row[env.stateDim*2+2] = 0
				}
			}
		}
		allTransTemp[i-1] = row
	}

	var allTrans [][]float64
	if env.EnsembleSeed != 0 {
		tempRnd := rand.New(rand.NewSource(int64(env.EnsembleSeed)))
		filteredLen := int(float64(len(allTransTemp)) * (1 - env.DropPerc))
		filteredIdx := tempRnd.Perm(len(allTransTemp))[:filteredLen]
		allTrans = make([][]float64, filteredLen)
		for i := 0; i < filteredLen; i++ {
			allTrans[i] = allTransTemp[filteredIdx[i]]
		}
	} else {
		allTrans = allTransTemp
	}

	env.offlineData = allTrans
	env.offlineStarts = env.SearchOfflineStart(allTrans)
	env.offlineModel = transModel.New(env.NumberOfActions, env.stateDim, "euclidean")
	env.offlineModel.BuildTree(allTrans)

	pdf := make([]float64, env.Neighbor_num)
	temp := 0.0
	for i := 0; i < env.Neighbor_num; i++ {
		pdf[i] = math.Pow(0.5, float64(i+1))
		temp += pdf[i]
	}
	for i := 0; i < env.Neighbor_num; i++ {
		pdf[i] = pdf[i] / temp
	}
	env.neighbor_prob = make([]float64, env.Neighbor_num)
	temp1 := 0.0
	for i := 0; i < env.Neighbor_num; i++ {
		env.neighbor_prob[i] = temp1 + pdf[i]
		temp1 = env.neighbor_prob[i]
	}
	return nil
}

func (env *KnnModelEnv) SearchOfflineStart(allTrans [][]float64) []int {
	starts := []int{0}
	for i := 0; i < len(allTrans)-1; i++ { // not include the end of run
		if allTrans[i][len(allTrans[i])-1] == 1 {
			starts = append(starts, i+1)
		}
	}
	return starts
}

func (env *KnnModelEnv) randomizeState() rlglue.State {
	randIdx := env.rng.Intn(len(env.offlineStarts))
	state := env.offlineData[env.offlineStarts[randIdx]][:env.stateDim]
	return state
}

// Start returns an initial observation.
func (env *KnnModelEnv) Start() rlglue.State {
	env.state = env.randomizeState()
	state_copy := make([]float64, env.stateDim)
	copy(state_copy, env.state)
	return state_copy
}

func (env *KnnModelEnv) Step(act rlglue.Action) (rlglue.State, float64, bool) {
	//fmt.Println("---------------------")
	actInt, _ := tpo.GetInt(act)
	_, nextStates, rewards, terminals, distances := env.offlineModel.SearchTree(env.state, actInt, env.Neighbor_num)

	sum := 0.0
	for i := 0; i < len(distances); i++ {
		sum += distances[i] + math.Pow(10, -6)
	}
	pdf := make([]float64, len(distances))
	for i := 0; i < len(distances); i++ {
		pdf[i] = 1.0 - ((distances[i] + math.Pow(10, -6)) / sum)
	}

	normalizedPdf := make([]float64, len(distances))
	normalizedSum := 0.0
	for i := 0; i < len(distances); i++ {
		normalizedSum += pdf[i]
	}
	for i := 0; i < len(distances); i++ {
		normalizedPdf[i] = (pdf[i] / normalizedSum)
	}

	env.neighbor_prob = make([]float64, len(distances))
	temp1 := 0.0
	for i := 0; i < len(distances); i++ {
		env.neighbor_prob[i] = temp1 + normalizedPdf[i]
		temp1 = env.neighbor_prob[i]
	}

	//fmt.Println("Distances: ", distances)
	//fmt.Println("PDF: ", pdf)
	//fmt.Println("Normalized PDF: ", normalizedPdf)
	//fmt.Println("CDF: ", env.neighbor_prob)
	//fmt.Println("Current State: ", env.state)
	//fmt.Println("Neighbour states: ", neighbourStates)
	//fmt.Println(nextStates)
	chosen := random.FreqSample(env.neighbor_prob)
	//fmt.Println(chosen)
	env.state = nextStates[chosen]
	var done bool
	if terminals[chosen] == 0 {
		done = false
	} else {
		done = true
	}

	state_copy := make([]float64, env.stateDim)
	copy(state_copy, env.state)
	return state_copy, rewards[chosen], done
}

//GetAttributes returns attributes for this environment.
func (env *KnnModelEnv) GetAttributes() rlglue.Attributes {
	// Add elements to attributes.
	attributes := struct {
		NumAction  int       `json:"numberOfActions"`
		StateDim   int       `json:"stateDimension"`
		StateRange []float64 `json:"stateRange"`
	}{
		env.NumberOfActions,
		env.stateDim,
		env.stateRange,
	}
	attr, err := json.Marshal(&attributes)
	if err != nil {
		env.Message("err", "environment.knnModel could not Marshal its JSON attributes: "+err.Error())
	}
	return attr
}

func (env *KnnModelEnv) Distance(state1 rlglue.State, state2 rlglue.State) float64 {
	var distance float64
	var squareddistance float64
	squareddistance = 0.0
	for i := 0; i < len(state1); i++ {
		squareddistance += math.Pow(state1[i]-state2[i], 2)
	}
	distance = math.Pow(squareddistance, 0.5)
	return distance
}