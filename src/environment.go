package polecalc

import (
	"fmt"
	"os"
	"json"
	"bytes"
	"io/ioutil"
	"reflect"
	"math"
)

// Holds all the necessary data for evaluating functions in the cuprate system
type Environment struct {
	// program parameters
	GridLength  uint32  // points per side in Brillouin zone; typical value ~ 64
	ImGc0Bins   uint    // number of bins to use when calculating the imaginary part of the electron Green's function
	ReGc0Points uint    // number of points to use on each side of the 1/x singularity when calculating ReGc0
	ReGc0dw     float64 // distance away from the singularity to step when calculating ReGc0
	NumProcs    uint16  // number of processes to use for mesh functions
	InitD1,     // initial values for self-consistent parameters
	InitMu,
	InitF0 float64

	// system constant physical parameters
	Alpha int8 // either -1 (d-wave) or +1 (s-wave)
	T,    // hopping energy for the physical electron
	T0, // Overall energy scale (default = 1.0)
	Tz, // z-direction hopping energy (|tz| < 0.3 or so)
	Thp, // Diagonal (next-nearest-neighbor) hopping energy (similar range as tz)
	X, // Doping / holon excess (0 < x < ~0.2)
	DeltaS, // spin gap
	CS float64 // coefficient for k deviation in omega_q

	// self-consistently determined physical parameters
	D1, // diagonal hopping parameter generated by two-hole process
	Mu, // holon chemical potential
	F0 float64 // superconducting order parameter

	// cached value: must be reset with EpsilonMin() if D1 changes
	EpsilonMin float64
}

// The one-holon hopping energy Th is determined by Environment parameters
func (env *Environment) Th() float64 {
	return env.T0 * (1 - env.X)
}

// Spinon chemical potential
func (env *Environment) Lambda() float64 {
	return math.Sqrt(math.Pow(env.DeltaS, 2.0) + math.Pow(env.CS, 2.0))
}

// Set self-consistent parameters to the initial values as specified by the Environment
func (env *Environment) Initialize() {
	// hard-coded defaults
	if env.NumProcs <= 0 {
		env.NumProcs = 1
	}
	// specified defaults
	env.D1 = env.InitD1
	env.Mu = env.InitMu
	env.F0 = env.InitF0
	// must be determined after system is otherwise initialized
	env.EpsilonMin = EpsilonMin(*env)
}

func (env *Environment) ZeroTempErrors() string {
	return fmt.Sprintf("errors - d1: %f; mu: %f; f0: %f", ZeroTempD1AbsError(*env), ZeroTempMuAbsError(*env), ZeroTempF0AbsError(*env))
}

// Convert the Environment to string by returning a JSON representation
func (env *Environment) String() string {
	envBytes, err := json.Marshal(env)
	if err != nil {
		return ""
	}
	envStr := bytes.NewBuffer(envBytes).String()
	return envStr
}

// Construct an Environment from the JSON file with given path.
// Self-consistent parameters are not set to values given by Init fields.
func EnvironmentFromFile(filePath string) (*Environment, os.Error) {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return EnvironmentFromBytes(fileContents)
}

// Convert to string and pass to EnvironmentFromBytes
func EnvironmentFromString(jsonData string) (*Environment, os.Error) {
	jsonBytes, err := StringToBytes(jsonData)
	if err != nil {
		return nil, err
	}
	return EnvironmentFromBytes(jsonBytes)
}

// Construct an Environment from the given JSON byte slice.
// Self-consistent parameters are not set to values given by Init fields.
func EnvironmentFromBytes(jsonData []byte) (*Environment, os.Error) {
	jsonObject := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &jsonObject); err != nil {
		return nil, err
	}
	return EnvironmentFromObject(jsonObject)
}

// Construct an Environment from the given JSON object.
// Self-consistent parameters are not set to values given by Init fields.
func EnvironmentFromObject(jsonObject map[string]interface{}) (*Environment, os.Error) {
	env := new(Environment)
	envValue := reflect.Indirect(reflect.ValueOf(env))
	for key, value := range jsonObject {
		field := envValue.FieldByName(key)
		fieldType := field.Type().Name()
		// Hack to get around Unmarshal treating all numbers as 
		// float64's. Will need to extend this for other types if they 
		// show up in Environment (or come up with a more clever 
		// solution).
		if fieldType == "uint" {
			field.Set(reflect.ValueOf(uint(value.(float64))))
		} else if fieldType == "uint16" {
			field.Set(reflect.ValueOf(uint16(value.(float64))))
		} else if fieldType == "uint32" {
			field.Set(reflect.ValueOf(uint32(value.(float64))))
		} else if fieldType == "int8" {
			field.Set(reflect.ValueOf(int8(value.(float64))))
		} else {
			field.Set(reflect.ValueOf(value))
		}
	}
	return env, nil
}

// Write the Environment to a JSON file at the given path
func (env *Environment) WriteToFile(filePath string) os.Error {
	if err := WriteToJSONFile(env, filePath); err != nil {
		return err
	}
	return nil
}
