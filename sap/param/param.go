package param

/*
Parameters are tunable switches on the system that are tuned in very specific
ways.

Each tunable parameter is able to inspect the system to determine the current
value, calculate a new value according to system environment and/or user input,
and be able to apply the new value upon request.

There's only one way to tune a parameter, however a parameter can be referred
to by one or more SAP notes.
*/

/*
Parameter is a tunable parameter, usually calculated based on user input and/or automatically.
Parameter is immutable. Internal state changes can only be made to copies.
*/
type Parameter interface {
	Inspect() (Parameter, error)                             // Read the parameter values from current system.
	Optimise(additionalInput interface{}) (Parameter, error) // Calculate new values based on internal states, and return a copy of new states.
	Apply(additionalInput interface{}) error                 // Apply the parameter value without further calculation.
}
