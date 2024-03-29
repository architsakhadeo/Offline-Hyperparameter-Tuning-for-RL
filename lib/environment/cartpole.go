package environment

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"

	"github.com/stellentus/cartpoles/lib/logger"
	"github.com/stellentus/cartpoles/lib/rlglue"
)

const (
	gravity               = 9.8
	masscart              = 1.0
	masspole              = 0.1
	totalMass             = (masspole + masscart)
	length                = 0.5 // half-length
	polemassLength        = (masspole * length)
	forceMag              = 10.0
	tau                   = 0.02
	thetaRhresholdRadians = (12 * 2 * math.Pi / 360)
	xThreshold            = 2.4
	stateMin              = -0.05
	stateMax              = 0.05
)

type CartpoleSettings struct {
	Seed               int64     `json:"seed"`
	Delays             []int     `json:"delays"`
	PercentNoiseState  []float64 `json:"percent_noise"`
	PercentNoiseAction float64   `json:"percent_noise_action"`
	PositionRwd        bool      `json:"rwd_position"`
	OneRwd             bool      `json:"rwd_one"`
	//RandomizeStartState bool      `json:"randomize_start_state"`
}

type Cartpole struct {
	logger.Debug
	CartpoleSettings
	state             rlglue.State
	rng               *rand.Rand
	rngStartState     *rand.Rand
	buffer            [][]float64
	bufferInsertIndex []int
}

func init() {
	Add("cartpole", NewCartpole)
}

func NewCartpole(logger logger.Debug) (rlglue.Environment, error) {
	return &Cartpole{Debug: logger}, nil
}

// Initialize configures the environment with the provided parameters and resets any internal state.
func (env *Cartpole) Initialize(run uint, attr rlglue.Attributes) error {
	set := CartpoleSettings{}
	err := json.Unmarshal(attr, &set)
	if err != nil {
		err = errors.New("environment.Cartpole settings error: " + err.Error())
		env.Message("err", err)
		return err
	}
	set.Seed += int64(run)

	return env.InitializeWithSettings(set)
}

func (env *Cartpole) InitializeWithSettings(set CartpoleSettings) error {
	env.CartpoleSettings = set
	//fmt.Println("Env Seed: ", env.Seed)
	//fmt.Println("Env CartpoleSettings Seed: ", env.CartpoleSettings.Seed)
	//fmt.Println("Set Seed:", set.Seed)
	//fmt.Println("Seed actually used by the environment: ", env.Seed)
	env.rng = rand.New(rand.NewSource(env.Seed)) // Create a new rand source for reproducibility
	env.rngStartState = rand.New(rand.NewSource(env.Seed))

	if len(env.PercentNoiseState) == 1 {
		// Copy it for all dimensions
		noise := env.PercentNoiseState[0]
		env.PercentNoiseState = []float64{noise, noise, noise, noise}
	} else if len(env.PercentNoiseState) != 4 && len(env.PercentNoiseState) != 0 {
		err := fmt.Errorf("environment.Cartpole requires percent_noise to be length 4, 1, or 0, not length %d", len(env.PercentNoiseState))
		env.Message("err", err)
		return err
	}

	if len(env.Delays) == 1 {
		// Copy it for all dimensions
		delay := env.Delays[0]
		env.Delays = []int{delay, delay, delay, delay}
	} else if len(env.Delays) != 4 && len(env.Delays) != 0 {
		err := fmt.Errorf("environment.Cartpole requires delays to be length 4, 1, or 0, not length %d", len(env.Delays))
		env.Message("err", err)
		return err
	}

	// If noise is off, set array to nil
	totalNoise := 0.0
	for _, noise := range env.PercentNoiseState {
		totalNoise += noise
	}
	if totalNoise == 0.0 {
		env.PercentNoiseState = nil
	}

	// If delays are off, set array to nil
	totalDelay := 0
	for _, noise := range env.Delays {
		totalDelay += noise
	}
	if totalDelay == 0 {
		env.Delays = nil
	}

	env.state = make(rlglue.State, 4)

	if len(env.Delays) != 0 {
		env.buffer = make([][]float64, 4)
		for i := range env.buffer {
			env.buffer[i] = make([]float64, env.Delays[i])
		}
		env.bufferInsertIndex = make([]int, 4)
	}

	env.Message("cartpole settings", fmt.Sprintf("%+v", env.CartpoleSettings))

	return nil
}

func (env *Cartpole) noisyState(startS bool) rlglue.State {
	stateLowerBound := []float64{-2.4, -4.0, -(12 * 2 * math.Pi / 360), -3.5}
	stateUpperBound := []float64{2.4, 4.0, (12 * 2 * math.Pi / 360), 3.5}

	state := make(rlglue.State, 4)
	copy(state, env.state)

	if len(env.PercentNoiseState) != 0 {
		// Only add noise if it's configured
		for i := range state {
			state[i] += env.randFloat(env.PercentNoiseState[i]*stateLowerBound[i], env.PercentNoiseState[i]*stateUpperBound[i], startS)
			state[i] = env.clamp(state[i], stateLowerBound[i], stateUpperBound[i])
		}
	}
	//if startS == true {
	//	fmt.Println("Randomize Noisy Start State: ", state)
	//}
	return state
}

func (env *Cartpole) clamp(x float64, min float64, max float64) float64 {
	return math.Max(min, math.Min(x, max))
}

func (env *Cartpole) randomizeState(randomizeStartStateCondition bool, startS bool) {
	for i := range env.state {
		env.state[i] = env.randFloat(stateMin, stateMax, startS)
	}

	if randomizeStartStateCondition == true {
		env.state[0] = env.randFloat(-xThreshold/3.0, xThreshold/3.0, startS)
		env.state[2] = env.randFloat(-thetaRhresholdRadians/3.0, thetaRhresholdRadians/3.0, startS)
	}
	//fmt.Println("Randomize Start State: ", env.state)
}

func (env *Cartpole) randFloat(min, max float64, startS bool) float64 {
	if startS == true {
		return env.rngStartState.Float64()*(max-min) + min
	} else {
		return env.rng.Float64()*(max-min) + min
	}
}

func (env *Cartpole) failureType() string {
	var info string
	x, theta := env.state[0], env.state[2]
	if (x < -xThreshold) || (x > xThreshold) {
		info = "pos"
	} else if (theta < -thetaRhresholdRadians) || (theta > thetaRhresholdRadians) {
		info = "ang"
	} else {
		info = ""
	}
	return info
}

// Start returns an initial observation.
func (env *Cartpole) Start(randomizeStartStateCondition bool) (rlglue.State, string) {
	env.randomizeState(randomizeStartStateCondition, true)
	return env.getObservations(true), env.failureType()
}

// Step takes an action and provides the resulting reward, the new observation, and whether the state is terminal.
// For this continuous environment, it's only terminal if the action was invalid.
func (env *Cartpole) Step(act rlglue.Action, randomizeStartStateCondition bool) (rlglue.State, float64, bool, string) {
	x, xDot, theta, thetaDot := env.state[0], env.state[1], env.state[2], env.state[3]

	force := forceMag
	if env.PercentNoiseAction != 0 {
		// Only add noise if it's configured
		//force += env.randFloat(0, force*env.PercentNoiseAction)
		force += env.rng.NormFloat64() * (env.PercentNoiseAction * forceMag)
	}
	//fmt.Println(force, env.rng.NormFloat64() * (env.PercentNoiseAction * forceMag))
	//fmt.Println("")

	// If the action is anything other than exactly 0, consider the action to have been the "move right" action.
	if act == 0 {
		force *= -1
	}

	costheta := math.Cos(theta)
	sintheta := math.Sin(theta)

	temp := (force + polemassLength*thetaDot*thetaDot*sintheta) / totalMass
	thetaacc := (gravity*sintheta - costheta*temp) / (length * (4.0/3.0 - masspole*costheta*costheta/totalMass))
	xacc := temp - polemassLength*thetaacc*costheta/totalMass

	// euler
	x = x + tau*xDot
	xDot = xDot + tau*xacc
	theta = theta + tau*thetaDot
	thetaDot = thetaDot + tau*thetaacc

	env.state = rlglue.State{x, xDot, theta, thetaDot}

	done := (x < -xThreshold) || (x > xThreshold) || (theta < -thetaRhresholdRadians) || (theta > thetaRhresholdRadians)
	var info string
	info = env.failureType()

	var reward float64
	if env.CartpoleSettings.PositionRwd {
		reward = -(math.Abs(theta)) / thetaRhresholdRadians
		if (x < -xThreshold) || (x > xThreshold) {
			//fmt.Println(reward, done)
			reward -= 1
		}
		//fmt.Println("---", reward, done)
	} else if env.CartpoleSettings.OneRwd && (!done) {
		reward = 1
	}
	if done {
		if (!env.CartpoleSettings.PositionRwd) && (!env.CartpoleSettings.OneRwd) {
			reward = -1.0
		}
		env.randomizeState(randomizeStartStateCondition, true)
		return env.getObservations(true), reward, done, info
		//fmt.Println("HERE", theta, reward)
	}

	return env.getObservations(false), reward, done, info
}

func (env *Cartpole) getObservations(startS bool) rlglue.State {
	// Add noise to state to get observations
	observations := env.noisyState(startS)

	// Add delays
	if len(env.Delays) != 0 {
		for i, obs := range observations {
			observations[i] = env.buffer[i][env.bufferInsertIndex[i]]                 // Load the delayed observation
			env.buffer[i][env.bufferInsertIndex[i]] = obs                             // Store the current state
			env.bufferInsertIndex[i] = (env.bufferInsertIndex[i] + 1) % env.Delays[i] // Update the insertion point
		}
	}
	return observations
}

// GetAttributes returns attributes for this environment.
func (env *Cartpole) GetAttributes() rlglue.Attributes {
	// Add elements to attributes.
	attributes := struct {
		NumAction  int       `json:"numberOfActions"`
		StateDim   int       `json:"stateDimension"`
		StateRange []float64 `json:"stateRange"`
	}{
		2,
		4,
		[]float64{4.8, 8.0, (2 * 12 * 2 * math.Pi / 360), 7.0},
	}

	attr, err := json.Marshal(&attributes)
	if err != nil {
		env.Message("err", "environment.Cartpole could not Marshal its JSON attributes: "+err.Error())
	}

	return attr
}

func (env *Cartpole) GetInfo(info string, value float64) interface{} {
	return nil
}
