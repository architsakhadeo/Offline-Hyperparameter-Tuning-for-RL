package logger

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/stellentus/cartpoles/lib/rlglue"
)

type DataConfig struct {
	// ShouldLogTraces determines whether traces are saved (reward, state, prevState, action, and any
	// other provided values). Rewards are always recorded.
	ShouldLogTraces bool

	// ShouldLogLearnProg determines whether learning progress are saved. The learning progress can be
	// mean squared TD error.
	ShouldLogLearnProg bool

	// CacheTracesInRAM determines whether traces are kept in RAM.
	CacheTracesInRAM bool

	// ShouldLogEpisodeLengths determines whether episode lengths are saved.
	ShouldLogEpisodeLengths bool

	// ShouldLogReturns determines whether episode returns are saved.
	ShouldLogReturns bool

	// ShouldLogRewards determines whether rewards are saved.
	ShouldLogRewards bool

	// ShouldLogTotals determines whether the episode's total rewards and episodes are saved.
	ShouldLogTotals bool

	// ShouldLog determines whether rewards are saved.
	ShouldLog bool

	// BasePath is the path at which files are saved. A filename will be automatically set (rewards, traces, and episodes).
	// If not set, no file is saved.
	BasePath string

	// FileSuffix is a suffix for the filename (most often used for run numbers). A dash will automatically be inserted.
	FileSuffix string
}

type DataLogger struct {
	Debug

	DataConfig

	episodeLengths []int
	episodeReturns []float64
	rewards        []float64

	totalReward   float64
	totalEpisodes int

	// These are used if the trace is cached in RAM.
	prevState []rlglue.State
	currState []rlglue.State
	actions   []rlglue.Action
	terminals []int
	others    [][]float64
	learnProg []string
	info      []string

	// file is used for writing out the trace.
	file *os.File
	*bufio.Writer
}

// NewDataWithExtraVariables creates a new Data with extra headers. The length of headers
// should match the length of 'others' in every call to LogStepMulti.
func NewDataWithExtraVariables(debug Debug, config DataConfig, headers ...string) (*DataLogger, error) {
	lg := &DataLogger{
		Debug:      debug,
		DataConfig: config,
	}

	if lg.CacheTracesInRAM {
		lg.prevState = []rlglue.State{}
		lg.currState = []rlglue.State{}
		lg.actions = []rlglue.Action{}
		for _ = range headers {
			lg.others = append(lg.others, []float64{})
		}
	}

	if lg.ShouldLogEpisodeLengths {
		lg.episodeLengths = []int{}
	}

	if lg.ShouldLogReturns {
		lg.episodeReturns = []float64{}
	}

	if lg.ShouldLogRewards || lg.ShouldLogReturns || lg.ShouldLogTraces {
		lg.rewards = []float64{}
	}

	if lg.ShouldLogLearnProg {
		lg.learnProg = []string{}
	}

	if lg.BasePath == "" {
		// No files are being saved, so nothing else to do
		return lg, nil
	}

	// Ensure the directory exists (TODO ensure it's writable)
	err := os.MkdirAll(lg.BasePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	if !lg.ShouldLogTraces {
		// The trace is not being saved, so nothing else to do
		return lg, nil
	}

	lg.file, err = os.Create(path.Join(lg.BasePath, "traces-"+lg.FileSuffix+".csv"))
	if err != nil {
		return nil, err
	}
	err = lg.writeTraceHeader(headers...)
	if err != nil {
		return nil, err
	}
	lg.Writer = bufio.NewWriterSize(lg.file, 64*1024)

	return lg, nil
}

func NewData(debug Debug, config DataConfig) (*DataLogger, error) {
	return NewDataWithExtraVariables(debug, config)
}

func (lg *DataLogger) NumberOfEpisodes() int {
	return len(lg.episodeLengths)
}

func (lg *DataLogger) writeTraceHeader(headers ...string) error {
	// Write header row
	str := "new state,previous state,action,reward,terminal,info"
	for _, hdr := range headers {
		str += "," + hdr
	}
	str += "\n"
	_, err := lg.file.WriteString(str)
	return err
}

func (lg *DataLogger) RewardSince(step int) float64 {
	if !lg.ShouldLogRewards && !lg.ShouldLogReturns && !lg.ShouldLogTraces {
		return 0
	}

	var sum float64
	end := len(lg.rewards)
	for i := step; i < end; i++ {
		sum += lg.rewards[i]
	}
	return sum
}

func (lg *DataLogger) LogEpisodeNumber() {
	lg.totalEpisodes++
}

// LogEpisodeLength adds the provided episode length to the episode length log.
func (lg *DataLogger) LogEpisodeLength(steps int) {
	if !lg.ShouldLogEpisodeLengths {
		return
	}
	lg.episodeLengths = append(lg.episodeLengths, steps)
}

// LogEpisodeLength adds the provided episode length to the episode length log.
func (lg *DataLogger) LogEpisodeReturn(return_ float64) {
	if !lg.ShouldLogReturns {
		return
	}
	lg.episodeReturns = append(lg.episodeReturns, return_)
}

// LogStep adds information from a step to the step log. It must contain previous state, current state,
// and reward. It can optionally add other float64 values to be logged. (If so, LogStepHeader must be
// called to provide headers and so the logger knows how many to expect.)
func (lg *DataLogger) LogStep(prevState, currState rlglue.State, action rlglue.Action, reward float64, terminal bool, info string) {
	str := lg.logStep(prevState, currState, action, reward, terminal, info)
	if lg.ShouldLogTraces {
		_, err := lg.Writer.WriteString(str + "\n")
		if err != nil {
			lg.Debug.Error(&err)
		}
	}
}

// LogStepMulti is like LogStep, but it can optionally add other float64 values to be logged. (If so,
// LogStepHeader must be called to provide headers and so the logger knows how many to expect.)
func (lg *DataLogger) LogStepMulti(prevState, currState rlglue.State, action rlglue.Action, reward float64, terminal bool, info string, others ...float64) {
	str := lg.logStep(prevState, currState, action, reward, terminal, info)

	if lg.CacheTracesInRAM {
		for i, other := range others {
			lg.others[i] = append(lg.others[i], other)
		}
	}

	if lg.ShouldLogTraces {
		for _, val := range others {
			str += fmt.Sprintf(",%f", val)
		}
		_, err := lg.Writer.WriteString(str + "\n")
		if err != nil {
			lg.Debug.Error(&err)
		}
	}
}

// LogLearnProg add information about learning progress. The information could be MSTDE for batch RL.
func (lg *DataLogger) LogLearnProg(progress string) {
	if !lg.ShouldLogLearnProg {
		return
	}
	lg.learnProg = append(lg.learnProg, progress)
}

func (lg *DataLogger) logStep(prevState, currState rlglue.State, action rlglue.Action, reward float64, terminal bool, info string) string {
	if lg.ShouldLogRewards || lg.ShouldLogReturns {
		lg.rewards = append(lg.rewards, reward)
	}

	lg.totalReward += reward

	var termInt int
	if terminal {
		termInt = 1
	} else {
		termInt = 0
	}
	if lg.CacheTracesInRAM {
		lg.prevState = append(lg.prevState, prevState)
		lg.currState = append(lg.currState, currState)
		lg.actions = append(lg.actions, action)
		lg.terminals = append(lg.terminals, termInt)
		lg.info = append(lg.info, info)
	}

	if !lg.ShouldLogTraces {
		return ""
	}

	return fmt.Sprintf("%v,%v,%d,%f,%d,%s", currState, prevState, action, reward, termInt, info)
}

// Save persists the logged information to disk.
func (lg *DataLogger) SaveLog() error {
	if lg.BasePath == "" {
		return nil
	}

	if lg.ShouldLogTotals {
		file, err := os.Create(path.Join(lg.BasePath, "totals-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()

		// Write header row
		_, err = file.WriteString("total reward, total episodes\n")
		if err != nil {
			return err
		}
		// Write totals
		_, err = file.WriteString(fmt.Sprintf("%f,%d\n", lg.totalReward, lg.totalEpisodes))
		if err != nil {
			return err
		}
	}

	if lg.ShouldLogEpisodeLengths {
		file, err := os.Create(path.Join(lg.BasePath, "rewards-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()

		// Write header row
		_, err = file.WriteString("rewards\n")
		if err != nil {
			return err
		}
		// Write remaining rows
		for _, rew := range lg.rewards {
			_, err = file.WriteString(fmt.Sprintf("%f\n", rew))
			if err != nil {
				return err
			}
		}
	}

	if lg.ShouldLogEpisodeLengths {
		file, err := os.Create(path.Join(lg.BasePath, "episodes-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()
		// Write header row
		_, err = file.WriteString("episode lengths\n")
		if err != nil {
			return err
		}
		// Write remaining rows
		for _, ep := range lg.episodeLengths {
			_, err = file.WriteString(fmt.Sprintf("%d\n", ep))
			if err != nil {
				return err
			}
		}
	}

	if lg.ShouldLogReturns {
		file, err := os.Create(path.Join(lg.BasePath, "returns-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()
		// Write header row
		_, err = file.WriteString("returns\n")
		if err != nil {
			return err
		}
		// Write remaining rows
		for _, ep := range lg.episodeReturns {
			_, err = file.WriteString(fmt.Sprintf("%f\n", ep))
			if err != nil {
				return err
			}
		}
	}

	if lg.ShouldLogLearnProg {
		file, err := os.Create(path.Join(lg.BasePath, "progs-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()

		// Write header row
		_, err = file.WriteString("learning progress\n")
		if err != nil {
			return err
		}

		// Write remaining rows
		for _, prg := range lg.learnProg {
			_, err = file.WriteString(fmt.Sprintf("%v\n", prg))
			if err != nil {
				return err
			}
		}
	}

	if lg.ShouldLogTraces {
		lg.Writer.Flush()
		lg.file.Close()
	}

	return nil
}

func (lg *DataLogger) GetBasePath() string {
	return path.Base(lg.BasePath)
}

func (lg *DataLogger) loadLog(pth string, suffix string, loadRewards, loadEpisodes, loadTraces bool) error {
	lg.DataConfig = DataConfig{
		ShouldLogTraces:         loadTraces,
		ShouldLogEpisodeLengths: loadEpisodes,
		ShouldLogRewards:        loadRewards,
		BasePath:                pth,
		FileSuffix:              suffix,
	}
	lg.episodeLengths = []int{}
	lg.rewards = []float64{}

	if loadRewards && !loadTraces { // If traces exists, don't bother with rewards
		file, err := os.Open(path.Join(lg.BasePath, "rewards-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if !scanner.Scan() {
			return errors.New("Reward file was empth at '" + lg.BasePath + "'")
		} // else assume header is correct

		for scanner.Scan() {
			var val float64
			_, err = fmt.Sscanf(scanner.Text(), "%f", &val)
			if err != nil {
				return err
			}
			lg.rewards = append(lg.rewards, val)
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	if loadEpisodes {
		file, err := os.Open(path.Join(lg.BasePath, "episodes-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if !scanner.Scan() {
			return errors.New("Reward file was empth at '" + lg.BasePath + "'")
		} // else assume header is correct

		for scanner.Scan() {
			var val int
			_, err = fmt.Sscanf(scanner.Text(), "%d", &val)
			if err != nil {
				return err
			}
			lg.episodeLengths = append(lg.episodeLengths, val)
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	if loadTraces {
		file, err := os.Open(path.Join(lg.BasePath, "traces.csv-"+lg.FileSuffix+".csv"))
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if !scanner.Scan() {
			return errors.New("Reward file was empth at '" + lg.BasePath + "'")
		}
		headers := strings.Split(scanner.Text(), ",")

		if err := scanner.Err(); err != nil {
			return err
		}

		numSteps := len(lg.rewards)
		if numSteps != len(lg.prevState) || numSteps != len(lg.currState) || numSteps != len(lg.actions) {
			return fmt.Errorf("Data file CSV unequal columns: %d %d %d %d", len(lg.currState), len(lg.prevState), len(lg.actions), numSteps)
		}

		numOthers := len(headers) - 4
		for i := 0; i < numOthers; i++ {
			if len(lg.others[i]) != numSteps {
				return fmt.Errorf("Data file CSV extra column %d has %d rows instead of %d", i, len(lg.others[i]), numSteps)
			}
		}
	}

	return nil
}

func parseState(str string) (rlglue.State, error) {
	str = strings.TrimSpace(str)
	if str[0] != '[' && str[len(str)-1] != ']' {
		return nil, errors.New("Could not parse state '" + str + "'")
	}
	state := rlglue.State{}
	values := strings.Split(str[1:len(str)-1], ",")
	for _, strVal := range values {
		if val, err := strconv.ParseFloat(strVal, 64); err != nil {
			fmt.Println("values", values)
			return nil, errors.New("Could not parse state value '" + strVal + "' from '" + str + "': " + err.Error())
		} else {
			state = append(state, val)
		}
	}
	return state, nil
}

func parseActionDefaultZero(str string) rlglue.Action {
	if val, err := strconv.Atoi(str); err != nil {
		return 0
	} else {
		return rlglue.Action(val)
	}
}

func parseFloatDefaultZero(str string) float64 {
	if val, err := strconv.ParseFloat(str, 64); err != nil {
		return 0
	} else {
		return val
	}
}
