package base

import (
	"fmt"
	"errors"
	"strings"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

func argsToDefs (args []*CXArgument, names []string, mod *CXModule, cxt *CXProgram) ([]*CXDefinition, error) {
	if len(names) == len(args) {
		defs := make([]*CXDefinition, 0)
		for i, arg := range args {
			defs = append(defs, &CXDefinition{
				Name: names[i],
				Typ: arg.Typ,
				Value: arg.Value,
				Module: mod,
				Context: cxt,
			})
			// defs[names[i]] = &CXDefinition{
			// 	Name: names[i],
			// 	Typ: arg.Typ,
			// 	Value: arg.Value,
			// 	Module: mod,
			// 	Context: cxt,
			// }
		}
		return defs, nil
	} else {
		return nil, errors.New("Not enough definition names provided")
	}
}

func PrintCallStack (callStack []*CXCall) {
	for i, call := range callStack {
		tabs := strings.Repeat("___", i)
		if tabs == "" {
			fmt.Printf("%sfn:%s ln:%d, \tlocals: ", tabs, call.Operator.Name, call.Line)
		} else {
			fmt.Printf("↓%sfn:%s ln:%d, \tlocals: ", tabs, call.Operator.Name, call.Line)
		}
		

		lenState := len(call.State)
		idx := 0
		for _, def := range call.State {
			if def.Name == "_" || (len(def.Name) > len(NON_ASSIGN_PREFIX) && def.Name[:len(NON_ASSIGN_PREFIX)] == NON_ASSIGN_PREFIX) {
				continue
			}
			var valI32 int32
			var valI64 int64
			var valF32 float32
			var valF64 float64
			switch def.Typ.Name {
			case "i32":
				encoder.DeserializeRaw(*def.Value, &valI32)
				if idx == lenState - 1 {
					fmt.Printf("%s: %d", def.Name, valI32)
				} else {
					fmt.Printf("%s: %d, ", def.Name, valI32)
				}
			case "i64":
				encoder.DeserializeRaw(*def.Value, &valI64)
				if idx == lenState - 1 {
					fmt.Printf("%s: %d", def.Name, valI64)
				} else {
					fmt.Printf("%s: %d, ", def.Name, valI64)
				}
			case "f32":
				encoder.DeserializeRaw(*def.Value, &valF32)
				if idx == lenState - 1 {
					fmt.Printf("%s: %f", def.Name, valF32)
				} else {
					fmt.Printf("%s: %f, ", def.Name, valF32)
				}
			case "f64":
				encoder.DeserializeRaw(*def.Value, &valF64)
				if idx == lenState - 1 {
					fmt.Printf("%s: %f", def.Name, valF64)
				} else {
					fmt.Printf("%s: %f, ", def.Name, valF64)
				}
			case "byte":
				if idx == lenState - 1 {
					fmt.Printf("%s: %d", def.Name, (*def.Value)[0])
				} else {
					fmt.Printf("%s: %d, ", def.Name, (*def.Value)[0])
				}
			case "[]byte":
				var val []byte
				encoder.DeserializeRaw(*def.Value, &val)
				if idx == lenState - 1 {
					fmt.Printf("%s: %v", def.Name, val)
				} else {
					fmt.Printf("%s: %v, ", def.Name, val)
				}
			case "[]i32":
				var val []int32
				encoder.DeserializeRaw(*def.Value, &val)
				if idx == lenState - 1 {
					fmt.Printf("%s: %v", def.Name, val)
				} else {
					fmt.Printf("%s: %v, ", def.Name, val)
				}
			case "[]i64":
				var val []int64
				encoder.DeserializeRaw(*def.Value, &val)
				if idx == lenState - 1 {
					fmt.Printf("%s: %v", def.Name, val)
				} else {
					fmt.Printf("%s: %v, ", def.Name, val)
				}
			case "[]f32":
				var val []float32
				encoder.DeserializeRaw(*def.Value, &val)
				if idx == lenState - 1 {
					fmt.Printf("%s: %v", def.Name, val)
				} else {
					fmt.Printf("%s: %v, ", def.Name, val)
				}
			case "[]f64":
				var val []float64
				encoder.DeserializeRaw(*def.Value, &val)
				if idx == lenState - 1 {
					fmt.Printf("%s: %v", def.Name, val)
				} else {
					fmt.Printf("%s: %v, ", def.Name, val)
				}
			}
			
			idx++
		}
		fmt.Println()
	}
}

func callsEqual (call1, call2 *CXCall) bool {
	if call1.Line != call2.Line ||
		len(call1.State) != len(call2.State) ||
		call1.Operator != call2.Operator ||
		call1.ReturnAddress != call2.ReturnAddress ||
		call1.Module != call2.Module {
		return false
	}

	for k, v := range call1.State {
		if call2.State[k] != v {
			return false
		}
	}

	return true
}

func saveStep (call *CXCall) {
	lenCallStack := len(call.Context.CallStack.Calls)
	newStep := MakeCallStack(lenCallStack)

	if len(call.Context.Steps) < 1 {
		// First call, copy everything
		for i, call := range call.Context.CallStack.Calls {
			newStep.Calls[i] = MakeCallCopy(call, call.Module, call.Context)
		}

		call.Context.Steps = append(call.Context.Steps, newStep)
		return
	}
	
	lastStep := call.Context.Steps[len(call.Context.Steps) - 1]
	lenLastStep := len(lastStep.Calls)
	
	smallerLen := 0
	if lenLastStep < lenCallStack {
		smallerLen = lenLastStep
	} else {
		smallerLen = lenCallStack
	}
	
	// Everytime a call changes, we need to make a hard copy of it
	// If the call doesn't change, we keep saving a pointer to it

	for i, call := range call.Context.CallStack.Calls[:smallerLen] {
		if callsEqual(call, lastStep.Calls[i]) {
			// if they are equal
			// append reference
			newStep.Calls[i] = lastStep.Calls[i]
		} else {
			newStep.Calls[i] = MakeCallCopy(call, call.Module, call.Context)
		}
	}

	// sizes can be different. if this is the case, we hard copy the rest
	for i, call := range call.Context.CallStack.Calls[smallerLen:] {
		newStep.Calls[i + smallerLen] = MakeCallCopy(call, call.Module, call.Context)
	}
	
	call.Context.Steps = append(call.Context.Steps, newStep)
	return
}

// It "un-runs" a program
func (cxt *CXProgram) Reset() {
	cxt.CallStack = MakeCallStack(0)
	cxt.Steps = make([]*CXCallStack, 0)
	//cxt.ProgramSteps = nil
}

func (cxt *CXProgram) ResetTo(stepNumber int) {
	// if no steps, we do nothing. the program will run from step 0
	if len(cxt.Steps) > 0 {
		if stepNumber > len(cxt.Steps) {
			stepNumber = len(cxt.Steps) - 1
		}
		reqStep := cxt.Steps[stepNumber]

		newStep := MakeCallStack(len(reqStep.Calls))
		
		var lastCall *CXCall
		for j, call := range reqStep.Calls {
			newCall := MakeCallCopy(call, call.Module, call.Context)
			newCall.ReturnAddress = lastCall
			lastCall = newCall
			newStep.Calls[j] = newCall
		}

		cxt.CallStack = newStep
		cxt.Steps = cxt.Steps[:stepNumber]
	}
}

func (cxt *CXProgram) UnRun (nCalls int) {
	if len(cxt.Steps) > 0 && nCalls > 0 {
		if nCalls > len(cxt.Steps) {
			nCalls = len(cxt.Steps) - 1
		}

		reqStep := cxt.Steps[len(cxt.Steps) - nCalls]

		newStep := MakeCallStack(len(reqStep.Calls))
		
		var lastCall *CXCall
		for j, call := range reqStep.Calls {
			newCall := MakeCallCopy(call, call.Module, call.Context)
			newCall.ReturnAddress = lastCall
			lastCall = newCall
			newStep.Calls[j] = newCall
		}

		cxt.CallStack = newStep
		cxt.Steps = cxt.Steps[:len(cxt.Steps) - nCalls]
	}
}

func replPrintEvaluation (arg *CXArgument) {
	fmt.Printf(">> ")
	switch arg.Typ.Name {
	case "str":
		fmt.Printf("%#v\n", string(*arg.Value))
	case "bool":
		var val int32
		encoder.DeserializeRaw(*arg.Value, &val)
		if val == 0 {
			fmt.Printf("false\n")
		} else {
			fmt.Printf("true\n")
		}
	case "byte":
		fmt.Printf("%#v\n", *arg.Value)
	case "i32":
		var val int32
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "i64":
		var val int64
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "f32":
		var val float32
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "f64":
		var val float64
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "[]byte":
		var val []byte
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "[]i32":
		var val []int32
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "[]i64":
		var val []int64
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "[]f32":
		var val []float32
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	case "[]f64":
		var val []float64
		encoder.DeserializeRaw(*arg.Value, &val)
		fmt.Printf("%#v\n", val)
	default:
		fmt.Printf("")
	}
}

func beforeDot (str string ) (beforeDot string, afterDot string) {
	beforeDot = ""
	afterDot = ""
	beforeDotCounter := 0
	foundDot := false
	for _, letter := range str {
		if letter != '.' {
			beforeDot = concat(beforeDot, string(letter))
		} else {
			foundDot = true
			break
		}
		beforeDotCounter++
	}
	if foundDot {
		afterDot = str[beforeDotCounter + 1:] // ignore the dot
	}
	
	return beforeDot, afterDot
}

func getIdentParts (str string) []string {
	var identParts []string
	before, after := beforeDot(str) // => "Math", "myPoint.x"
	identParts = append(identParts, before)
	for after != "" {
		before, after = beforeDot(after)
		identParts = append(identParts, before)
	}
	return identParts
}

func getAllocSize (typ string) int {
	// I need to do arrays too
	switch typ {
	case "byte":
		return 1
	case "i32":
		return 4
	case "i64":
		return 8
	case "f32":
		return 4
	case "f64":
		return 8
	default:
		return -1
	}
}

func (cxt *CXProgram) Compile () *CXProgram {
	allocs := make(map[string]*CXArgument, 0)
	heap := *cxt.Heap
	heapCounter := 0

	for _, mod := range cxt.Modules {
		for _, fn := range mod.Functions {

			// using struct fields
			for _, expr := range fn.Expressions {


				for i, outName := range expr.OutputNames {
					allocSize := getAllocSize(expr.Operator.Outputs[i].Typ.Name)
					outName.Offset = heapCounter
					heapCounter = heapCounter + allocSize
				}
				
				
				for argIdx, arg := range expr.Arguments {
					if arg.Typ.Name == "ident" {
						identParts := getIdentParts(string(*arg.Value))

						//var def *CXDefinition
						if len(identParts) == 1 {
							// it's a current module's definition
							// TODO: compile later
						} else {
							// is the first part of the identifier a module, if yes:
							if identMod, err := cxt.GetModule(identParts[0]); err == nil {
								// identParts[1] will always be a definition if before is a module
								if len(identParts) == 2 {
									// then we are referring to the struct itself or a normal ident
									// TODO: compile later
								} else {
									// we're referring to a struct field then
									if len(identParts) > 3 {
										// nested structs
										// TODO: compile later
									} else {
										if def, err := identMod.GetDefinition(concat(identParts[1], ".", identParts[2])); err == nil {
											var heapArg *CXArgument
											defSize := len(*def.Value)
											if found, ok := allocs[concat(identParts[0], ".", identParts[1], ".", identParts[2])]; ok {
												heapArg = found
											} else {
												heapArg = &CXArgument{
													Typ: def.Typ,
													//Value: arg.Value, //irrelevant now
													Offset: heapCounter,
													Size: defSize,
												}
												allocs[concat(identParts[0], ".", identParts[1], ".", identParts[2])] = heapArg
											}
											
											// replacing argument
											expr.Arguments[argIdx] = heapArg

											heap = append(heap, *def.Value...)
											heapCounter = heapCounter + defSize
										} else {
											// this means it's a local definition (or it doesn't exist, but Run() takes care of this)
											
										}
									}
								}
							} else {
								if len(identParts) > 2 {
									fmt.Println("identParts > 2")
									// nested structs
									// TODO: compile later
								} else {
									if def, err := mod.GetDefinition(concat(identParts[0], ".", identParts[1])); err == nil {
										//def = mod.Definitions[concat(identParts[0], ".", identParts[1])]
										var heapArg *CXArgument
										defSize := len(*def.Value)
										if found, ok := allocs[concat(mod.Name, ".", identParts[0], ".", identParts[1])]; ok {
											heapArg = found
										} else {
											heapArg = &CXArgument{
												Typ: def.Typ,
												Offset: heapCounter,
												Size: defSize,
											}
											// mod.Name identParts[0]
											allocs[concat(mod.Name, ".", identParts[0], ".", identParts[1])] = heapArg
										}
										
										// replacing argument
										expr.Arguments[argIdx] = heapArg

										heap = append(heap, *def.Value...)
										heapCounter = heapCounter + defSize
									}
								}
							}
						}

						// checking if first part is a module
						
					}
				}
			}
		}
	}
	cxt.Heap = &heap
	return cxt
}

func (cxt *CXProgram) Run (withDebug bool, nCalls int) {
	var callCounter int = 0
	// we are going to do this if the CallStack is empty
	if cxt.CallStack != nil && len(cxt.CallStack.Calls) > 0 {
		// we resume the program
		lastCall := cxt.CallStack.Calls[len(cxt.CallStack.Calls) - 1]
		
		lastCall.call(withDebug, nCalls, callCounter)
	} else {
		// initialization and checking
		if mod, err := cxt.SelectModule("main"); err == nil {
			if fn, err := mod.SelectFunction("main"); err == nil {
				// main function
				state := make([]*CXDefinition, 0)
				mainCall := MakeCall(fn, state, nil, mod, mod.Context)
				
				cxt.CallStack.Calls = append(cxt.CallStack.Calls, mainCall)

				mainCall.call(withDebug, nCalls, callCounter)
			}
			
		} else {
			fmt.Println(err)
		}
	}
}

func (call *CXCall) call(withDebug bool, nCalls, callCounter int) {
	//  add a counter here to pause
	if nCalls > 0 && callCounter >= nCalls {
		return
	}
	callCounter++

	saveStep(call)
	if withDebug {
		PrintCallStack(call.Context.CallStack.Calls)
	}
	
	if call.Line >= len(call.Operator.Expressions) {
		
		// if len(call.Operator.Expressions) < 1 {
		// 	panic(fmt.Sprintf("Calling function without expressions '%s'", call.Operator.Name))
		// }
		
		// popping the stack
		call.Context.CallStack.Calls = call.Context.CallStack.Calls[:len(call.Context.CallStack.Calls) - 1]
		//outName := call.Operator.Output.Name


		outNames := make([]string, len(call.Operator.Outputs))
		for i, out := range call.Operator.Outputs {
			outNames[i] = out.Name
		}

		// this one is for returning result
		//output := call.State[outName]

		outputs := make([]*CXDefinition, len(call.Operator.Outputs))

		for i, outName := range outNames {
			//fmt.Printf("%s : %v\n", outName, call.State[outName])
			//fmt.Println(call.State[outName])

			//outputs[i] = call.State[outName]
			for _, stateDef := range call.State {
				if stateDef.Name == outName {
					outputs[i] = stateDef
				}
			}

			// if call.State[outName] == nil {
			// 	call.Context.PrintProgram(false)
			// }
		}

		allOutsNil := true
		for _, out := range outputs {
			if out != nil {
				allOutsNil = false
				break
			}
		}

		if allOutsNil && len(outputs) > 0 {
			outNames := call.Operator.Expressions[len(call.Operator.Expressions) - 1].OutputNames
			for i, outName := range outNames {
				for _, stateDef := range call.State {
					if stateDef.Name == outName.Name {
						outputs[i] = stateDef
					}
				}
			}
		}

		for i, out := range outputs {
			if out.Typ.Name != call.Operator.Outputs[i].Typ.Name {
				panic(fmt.Sprintf("output var '%s' is of type '%s'; function '%s' requires output of type '%s'",
					out.Name, out.Typ.Name, call.Operator.Name, call.Operator.Outputs[i].Typ.Name))
			}
		}
		
		//checking if output var has the same type as the required output
		// if output.Typ.Name != call.Operator.Output.Typ.Name {
		// 	panic(fmt.Sprintf("output var '%s' is of type '%s'; function '%s' requires output of type '%s'",
		// 		output.Name, output.Typ.Name, call.Operator.Name, call.Operator.Output.Typ.Name))
		// }

		if call.ReturnAddress != nil {
			//returnName := call.ReturnAddress.Operator.Expressions[call.ReturnAddress.Line - 1].OutputName

			returnNames := make([]string, len(call.ReturnAddress.Operator.Expressions[call.ReturnAddress.Line - 1].OutputNames))
			for i, out := range call.ReturnAddress.Operator.Expressions[call.ReturnAddress.Line - 1].OutputNames {
				returnNames[i] = out.Name
			}

			if len(outputs) > 0 {
				//defs := make([]*CXDefinition, len(returnNames))
				for i, returnName := range returnNames {
					def := MakeDefinition(returnName, outputs[i].Value, outputs[i].Typ)
					def.Module = call.Module
					def.Context = call.Context

					
					//call.ReturnAddress.State[returnName] = def

					idx := -1
					for i, stateDef := range call.ReturnAddress.State {
						if stateDef.Name == returnName {
							idx = i
							break
						}
					}
					if idx < 0 {
						call.ReturnAddress.State = append(call.ReturnAddress.State, def)
					} else {
						call.ReturnAddress.State[idx] = def
					}
				}
			}//  else {
			// 	panic(fmt.Sprintf("Function '%s' didn't return anything", call.Operator.Name))
			// }

			// if output != nil {
			// 	def := MakeDefinition(returnName, output.Value, output.Typ)
			// 	def.Module = call.Module
			// 	def.Context = call.Context
			// 	call.ReturnAddress.State[returnName] = def
			// } else {
			// 	panic(fmt.Sprintf("Function '%s' couldn't return anything", call.Operator.Name))
			// }

			call.ReturnAddress.call(withDebug, nCalls, callCounter)
		} else {
			// no return address. should only be for main
			call.Context.Outputs = outputs
			//fmt.Printf("\nProgram's output:\n%v\n", output.Value)
		}
	} else {
		fn := call.Operator
		
		globals := call.Module.Definitions
		state := call.State

		if expr, err := fn.GetExpression(call.Line); err == nil {
			// getting arguments
			//outName := expr.OutputName
			argsRefs, _ := expr.GetArguments()

			argsCopy := make([]*CXArgument, len(argsRefs))
			argNames := make([]string, len(argsRefs))

			//call.Context.PrintProgram(false)

			for i, inp := range expr.Operator.Inputs {
				argNames[i] = inp.Name
			}
			
			// we are modifying by reference, we need to make copies
			for i := 0; i < len(argsRefs); i++ {
				if argsRefs[i].Offset > -1 {
					offset := argsRefs[i].Offset
					size := argsRefs[i].Size
					//var val []byte
					//encoder.DeserializeRaw((*call.Context.Heap)[offset:offset+size], &val)
					//argsRefs[i].Value = &val
					val := (*call.Context.Heap)[offset:offset+size]
					argsRefs[i].Value = &val
					continue
				}
				if argsRefs[i].Typ.Name == "ident" {
					lookingFor := string(*argsRefs[i].Value)

					// in here we could create an entry in the state map

					found := false
					var local *CXDefinition
					var global *CXDefinition
					for _, stateDef := range state {
						if stateDef.Name == lookingFor {
							found = true
							local = stateDef
						}
					}
					if !found {
						for _, globalDef := range globals {
							if globalDef.Name == lookingFor {
								global = globalDef
							}
						}
					}
					
					//local := state[lookingFor]
					//global := globals[lookingFor]


					// fmt.Printf("Here: %s\n", lookingFor)
					// fmt.Printf("Here2: %s\n", string(*argsRefs[i].Value))
					
					if (local == nil && global == nil) {
						panic(fmt.Sprintf("'%s' is undefined", lookingFor))
					}

					// giving priority to local var
					if local != nil {
						argsCopy[i] = MakeArgument(local.Value, local.Typ)
						//argsCopy[i] = MakeArgumentCopy(local)
					} else {
						argsCopy[i] = MakeArgument(global.Value, global.Typ)
						//argsCopy[i] = MakeArgumentCopy(global)
					}
				} else {
					argsCopy[i] = argsRefs[i]
				}

				// checking if arguments types match with expressions required types
				if len(expr.Operator.Inputs) > 0 && expr.Operator.Inputs[i].Typ.Name != argsCopy[i].Typ.Name {
					panic(fmt.Sprintf("%s, line #%d: %s argument #%d is type '%s'; expected type '%s'",
						fn.Name, call.Line, expr.Operator.Name, i, argsCopy[i].Typ.Name, expr.Operator.Inputs[i].Typ.Name))
				}
			}

			// checking if native or not
			var values []*CXArgument
			opName := expr.Operator.Name
			
			switch opName {
			case "eval":
				//
			case "evolve":
				fnName := string(*argsCopy[0].Value)
				fnBag := string(*argsCopy[1].Value)
				
				var inps []float64
				float64Size := encoder.Size(float64(0))
				for i := 0; i < len(*argsCopy[2].Value) / float64Size; i++ {
					var inp float64
					from := i * float64Size
					to := (i + 1) * float64Size
					encoder.DeserializeRaw((*argsCopy[2].Value)[from:to], &inp)
					inps = append(inps, inp)
				}
				
				var outs []float64
				for i := 0; i < len(*argsCopy[3].Value) / float64Size; i++ {
					var out float64
					from := i * float64Size
					to := (i + 1) * float64Size
					encoder.DeserializeRaw((*argsCopy[3].Value)[from:to], &out)
					outs = append(outs, out)
				}

				var numberExprs int32
				encoder.DeserializeRaw(*argsCopy[4].Value, &numberExprs)
				var iterations int32
				encoder.DeserializeRaw(*argsCopy[5].Value, &iterations)
				var epsilon float64
				encoder.DeserializeRaw(*argsCopy[6].Value, &epsilon)
				
				evolutionErr := call.Context.Evolve(fnName, fnBag, inps, outs, int(numberExprs), int(iterations), epsilon)

				// return 1 for true or something like that
				val := encoder.Serialize(evolutionErr)
				values = append(values, MakeArgument(&val, MakeType("f64")))

				// flow control
			case "goTo":
				values = append(values, goTo(call, argsCopy[0], argsCopy[1], argsCopy[2]))
				// I/O functions
			case "printBool":
				var val int32
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				if val == 0 {
					fmt.Println("false")
				} else {
					fmt.Println("true")
				}
				values = append(values, argsCopy[0])
			case "printStr":
				fmt.Println(string(*argsCopy[0].Value))
				values = append(values, argsCopy[0])
			case "printByte":
				fmt.Println((*argsCopy[0].Value)[0])
				values = append(values, argsCopy[0])
			case "printI32":
				var val int32
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
			case "printI64":
				var val int64
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
			case "printF32":
				var val float32
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
			case "printF64":
				var val float64
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
			case "printByteA":
				fmt.Println(*argsCopy[0].Value)
				values = append(values, argsCopy[0])
			case "printI32A":
				var val []int32
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
			case "printI64A":
				var val []int64
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
			case "printF32A":
				var val []float32
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
			case "printF64A":
				var val []float64
				encoder.DeserializeRaw(*argsCopy[0].Value, &val)
				fmt.Println(val)
				values = append(values, argsCopy[0])
				// identity functions
			case "idStr": values = append(values, argsCopy[0])
			case "idByte": values = append(values, argsCopy[0])
			case "idI32": values = append(values, argsCopy[0])
			case "idI64": values = append(values, argsCopy[0])
			case "idF32": values = append(values, argsCopy[0])
			case "idF64": values = append(values, argsCopy[0])
			case "idByteA": values = append(values, argsCopy[0])
			case "idI32A": values = append(values, argsCopy[0])
			case "idI64A": values = append(values, argsCopy[0])
			case "idF32A": values = append(values, argsCopy[0])
			case "idF64A": values = append(values, argsCopy[0])
				// cast functions
			case "byteAToStr": values = append(values, castToStr(argsCopy[0]))
				// case "byte": values = append(values, castByte(argsCopy[0]))
				// case "i32": values = append(values, castI32(argsCopy[0]))
			case "i64ToI32": values = append(values, castToI32(argsCopy[0]))
			case "f32ToI32": values = append(values, castToI32(argsCopy[0]))
			case "f64ToI32": values = append(values, castToI32(argsCopy[0]))
			case "i32ToI64": values = append(values, castToI64(argsCopy[0]))
			case "f32ToI64": values = append(values, castToI64(argsCopy[0]))
			case "f64ToI64": values = append(values, castToI64(argsCopy[0]))
			case "i32ToF32": values = append(values, castToF32(argsCopy[0]))
			case "i64ToF32": values = append(values, castToF32(argsCopy[0]))
			case "f64ToF32": values = append(values, castToF32(argsCopy[0]))
			case "i32ToF64": values = append(values, castToF64(argsCopy[0]))
			case "i64ToF64": values = append(values, castToF64(argsCopy[0]))
			case "f32ToF64": values = append(values, castToF64(argsCopy[0]))
			// case "f32": values = append(values, castToF32(argsCopy[0]))
			// case "f64": values = append(values, castToF64(argsCopy[0]))
			// case "[]byte": values = append(values, castByteA(argsCopy[0]))
			// case "[]i32": values = append(values, castI32A(argsCopy[0]))
			// case "[]i64": values = append(values, castI64A(argsCopy[0]))
			// case "[]f32": values = append(values, castF32A(argsCopy[0]))
			// case "[]f64": values = append(values, castF64A(argsCopy[0]))
				// relational operators
			case "ltI32":
				values = append(values, ltI32(argsCopy[0], argsCopy[1]))
			case "gtI32":
				values = append(values, gtI32(argsCopy[0], argsCopy[1]))
			case "eqI32":
				values = append(values, eqI32(argsCopy[0], argsCopy[1]))
			case "ltI64":
				values = append(values, ltI64(argsCopy[0], argsCopy[1]))

				// 1. Keep trying to solve compile
				// 2. Bootstrap native functions
				// 3. Maybe make everything be in a byte array
				// 4. Make CXGO call affordances
				// 5. Nested expressions
				// 6. Create a REPL
				
			case "gtI64":
				values = append(values, gtI64(argsCopy[0], argsCopy[1]))
			case "eqI64":
				values = append(values, eqI64(argsCopy[0], argsCopy[1]))
				// struct operations
			case "initDef":
				values = append(values, initDef(argsCopy[0]))
				// arithmetic functions
			case "addI32":
				values = append(values, addI32(argsCopy[0], argsCopy[1]))
			case "mulI32":
				values = append(values, mulI32(argsCopy[0], argsCopy[1]))
			case "subI32":
				values = append(values, subI32(argsCopy[0], argsCopy[1]))
			case "divI32":
				values = append(values, divI32(argsCopy[0], argsCopy[1]))
			case "addI64":
				values = append(values, addI64(argsCopy[0], argsCopy[1]))
			case "mulI64":
				values = append(values, mulI64(argsCopy[0], argsCopy[1]))
			case "subI64":
				values = append(values, subI64(argsCopy[0], argsCopy[1]))
			case "divI64":
				values = append(values, divI64(argsCopy[0], argsCopy[1]))
			case "addF32":
				values = append(values, addF32(argsCopy[0], argsCopy[1]))
			case "mulF32":
				values = append(values, mulF32(argsCopy[0], argsCopy[1]))
			case "subF32":
				values = append(values, subF32(argsCopy[0], argsCopy[1]))
			case "divF32":
				values = append(values, divF32(argsCopy[0], argsCopy[1]))
			case "addF64":
				values = append(values, addF64(argsCopy[0], argsCopy[1]))
			case "mulF64":
				values = append(values, mulF64(argsCopy[0], argsCopy[1]))
			case "subF64":
				values = append(values, subF64(argsCopy[0], argsCopy[1]))
			case "divF64":
				values = append(values, divF64(argsCopy[0], argsCopy[1]))
				// array functions
			case "readByteA":
				values = append(values, readByteA(argsCopy[0], argsCopy[1]))
			case "writeByteA":
				values = append(values, writeByteA(argsCopy[0], argsCopy[1], argsCopy[2]))
			case "readI32A":
				values = append(values, readI32A(argsCopy[0], argsCopy[1]))
			case "writeI32A":
				values = append(values, writeI32A(argsCopy[0], argsCopy[1], argsCopy[2]))
			case "readI64A":
				values = append(values, readI64A(argsCopy[0], argsCopy[1]))
			case "writeI64A":
				values = append(values, writeI64A(argsCopy[0], argsCopy[1], argsCopy[2]))
			case "readF32A":
				values = append(values, readF32A(argsCopy[0], argsCopy[1]))
			case "writeF32A":
				values = append(values, writeF32A(argsCopy[0], argsCopy[1], argsCopy[2]))
			case "readF64A":
				values = append(values, readF64A(argsCopy[0], argsCopy[1]))
			case "writeF64A":
				values = append(values, writeF64A(argsCopy[0], argsCopy[1], argsCopy[2]))
				/* Time functions */
			case "sleep":
				values = append(values, sleep(argsCopy[0]))
			case "":
			}
			if len(values) > 0 {
				// operator was a native function
				for i, outName := range expr.OutputNames {
					if withDebug {
						replPrintEvaluation(values[i])
					}
					def := MakeDefinition(outName.Name, values[i].Value, values[i].Typ)
					def.Module = call.Module
					def.Context = call.Context

					// if outName.Offset > -1 {
					// 	call.State[outName.Offset] = def
					// } else {
						
					// }
					
					found := false
					for i, stateDef := range call.State {
						if stateDef.Name == outName.Name {
							found = true
							call.State[i] = def
						}
					}
					if !found {
						call.State = append(call.State, def)
					}
					


					// if values[i].Offset > -1 {
					// 	for values[i].Value
						
					// 	(*call.Context.Heap)[values[i].Offset:values[i].Offset + values[i].Size]
					// }



					// if argsRefs[i].Offset > -1 {
					// 	offset := argsRefs[i].Offset
					// 	size := argsRefs[i].Size
					// 	//var val []byte
					// 	//encoder.DeserializeRaw((*call.Context.Heap)[offset:offset+size], &val)
					// 	//argsRefs[i].Value = &val
					// 	val := (*call.Context.Heap)[offset:offset+size]
					// 	argsRefs[i].Value = &val
					// 	continue
					// }
					
				}
				call.Line++
				call.call(withDebug, nCalls, callCounter)
			} else {
				// operator was not a native function
				call.Line++ // once the subcall finishes, call next line of the
				if argDefs, err := argsToDefs(argsCopy, argNames, call.Module, call.Context); err == nil {
					subcall := MakeCall(expr.Operator, argDefs, call, call.Module, call.Context)
					call.Context.CallStack.Calls = append(call.Context.CallStack.Calls, subcall)
					subcall.call(withDebug, nCalls, callCounter)
					return
				} else {
					fmt.Println(err)
				}
			}
		} else {
			fmt.Println(err)
		}
	}
}
