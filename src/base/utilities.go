package base

import (
	"fmt"
	"math/rand"
	"time"
	"bytes"
	"strings"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max - min) + min
}

func removeDuplicatesInt(elements []int) []int {
	// Use map to record duplicates as we find them.
	encountered := map[int]bool{}
	result := []int{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func removeDuplicates(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}

func concat (strs ...string) string {
	var buffer bytes.Buffer
	
	for i := 0; i < len(strs); i++ {
		buffer.WriteString(strs[i])
	}
	
	return buffer.String()
}



// (modules (main (definitions (str ))
//                (functions (= double (addF64 ((f64 n) (f64 n))))
//                           (= triple (subF64 ((f64 n) (f64 n))))))
// 	 (math (definitions)))

// (modules (module main
//                  (definitions
//                      (definition greeting str "hello")
//                      (definition ten i32 10)
//                    (definition ten i32 10)
//                      )))


func printValue (value *[]byte, typName string) string {
	var argName string
	switch typName {
	case "str":
		argName = fmt.Sprintf("\"%s\"", string(*value))
	case "bool":
		var val int32
		encoder.DeserializeRaw(*value, &val)
		if val == 0 {
			argName = "false"
		} else {
			fmt.Printf("true")
			argName = "true"
		}
	case "byte":
		argName = fmt.Sprintf("%#v", value)
	case "i32":
		var val int32
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "i64":
		var val int64
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "f32":
		var val float32
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "f64":
		var val float64
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "[]byte":
		var val []byte
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "[]i32":
		var val []int32
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "[]i64":
		var val []int64
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "[]f32":
		var val []float32
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	case "[]f64":
		var val []float64
		encoder.DeserializeRaw(*value, &val)
		argName = fmt.Sprintf("%#v", val)
	default:
		argName = string(*value)
	}

	return argName
}

func rep (str string, n int) string {
	return strings.Repeat(str, n)
}

// func (cxt *CXProgram) PrintProgram (isCompressed bool) {
// 	tab := "\t"
// 	nl := "\n"
// 	if isCompressed {
// 		tab = ""
// 		nl = ""
// 	}
	
// 	fmt.Println()
// 	fmt.Printf("(Modules %s", nl)
// 	for _, mod := range cxt.Modules {
// 		if mod.Name == CORE_MODULE {
// 			continue
// 		}
// 		fmt.Printf("%s(Module %s %s", rep(tab, 1), mod.Name, nl)

// 		fmt.Printf("%s(Imports %s", rep(tab, 2), nl)
// 		fmt.Printf("%s)%s", rep(tab, 2), nl) // imports
		
// 		fmt.Printf("%s(Definitions %s", rep(tab, 2), nl)

// 		for _, def := range mod.Definitions {
// 			fmt.Printf("%s(Definition %s %s %s)%s",
// 				rep(tab, 3),
// 				def.Name,
// 				def.Typ.Name,
// 				printValue(def.Value, def.Typ.Name),
// 				nl)
// 		}
		
// 		fmt.Printf("%s)%s", rep(tab, 2), nl) // definitions

// 		fmt.Printf("%s(Structs %s", rep(tab, 2), nl)

// 		for _, strct := range mod.Structs {
// 			fmt.Printf("%s(Struct %s", rep(tab, 3), nl)

// 			for _, fld := range strct.Fields {
// 				fmt.Printf("%s%s %s%s", rep(tab, 4), fld.Name, fld.Typ.Name, nl)
// 			}
			
// 			fmt.Printf("%s)%s", rep(tab, 3), nl) // structs
// 		}
		
// 		fmt.Printf("%s)%s", rep(tab, 2), nl) // structs

// 		fmt.Printf("%s(Functions %s", rep(tab, 2), nl)

// 		for _, fn := range mod.Functions {
// 			fmt.Printf("%s(Function %s%s", rep(tab, 3), fn.Name, nl)

// 			fmt.Printf("%s(Inputs %s", rep(tab, 4), nl)
// 			for _, inp := range fn.Inputs {
// 				fmt.Printf("%s(Input %s %s)%s", rep(tab, 5), inp.Name, inp.Typ.Name, nl)
// 			}
// 			fmt.Printf("%s)%s", rep(tab, 4), nl) // inputs

// 			fmt.Printf("%s(Outputs %s", rep(tab, 4), nl)
// 			for _, out := range fn.Outputs {
// 				fmt.Printf("%s(Output %s %s)%s", rep(tab, 5), out.Name, out.Typ.Name, nl)
// 			}
// 			fmt.Printf("%s)%s", rep(tab, 4), nl) // outputs

// 			fmt.Printf("%s(Expressions %s", rep(tab, 4), nl)
// 			for _, expr := range fn.Expressions {
// 				_ = expr
// 				fmt.Printf("%s(Expression %s", rep(tab, 5), nl)

// 				fmt.Printf("%s(Operator %s)%s", rep(tab, 6), expr.Operator.Name, nl)
				
// 				fmt.Printf("%s(OutputNames %s", rep(tab, 6), nl)
// 				for _, outName := range expr.OutputNames {
// 					fmt.Printf("%s(OutputName %s)%s", rep(tab, 7), outName.Name, nl)
// 				}
// 				fmt.Printf("%s)%s", rep(tab, 6), nl)
				
// 				fmt.Printf("%s(Arguments %s", rep(tab, 6), nl)
// 				for _, arg := range expr.Arguments {
// 					fmt.Printf("%s(Argument %s %s)%s", rep(tab, 7), printValue(arg.Value, arg.Typ.Name), arg.Typ.Name, nl)
// 				}
// 				fmt.Printf("%s)%s", rep(tab, 6), nl)
				
// 				fmt.Printf("%s)%s", rep(tab, 5), nl)
// 			}
// 			fmt.Printf("%s)%s", rep(tab, 4), nl) // expressions
			
// 			fmt.Printf("%s)%s", rep(tab, 3), nl) // function
// 		}
		
// 		fmt.Printf("%s)%s", rep(tab, 2), nl) // functions
		
// 		fmt.Printf("%s)%s", rep(tab, 1), nl) // modules
// 	}
// 	fmt.Printf(")")
// 	fmt.Println()
// }



func (cxt *CXProgram) PrintProgram(withAffs bool) {

	fmt.Println("Program")
	if withAffs {
		for i, aff := range cxt.GetAffordances() {
			fmt.Printf(" * %d.- %s\n", i, aff.Description)
		}
	}

	i := 0
	for _, mod := range cxt.Modules {
		if mod.Name == CORE_MODULE {
			continue
		}
		fmt.Printf("%d.- Module: %s\n", i, mod.Name)

		if withAffs {
			for i, aff := range mod.GetAffordances() {
				fmt.Printf("\t * %d.- %s\n", i, aff.Description)
			}
		}

		if len(mod.Imports) > 0 {
			fmt.Println("\tImports")
		}

		j := 0
		for _, imp := range mod.Imports {
			fmt.Printf("\t\t%d.- Import: %s\n", j, imp.Name)
			j++
		}

		if len(mod.Definitions) > 0 {
			fmt.Println("\tDefinitions")
		}

		j = 0
		for _, v := range mod.Definitions {
			fmt.Printf("\t\t%d.- Definition: %s %s\n", j, v.Name, v.Typ.Name)
			j++
		}

		if len(mod.Structs) > 0 {
			fmt.Println("\tStructs")
		}

		j = 0
		for _, strct := range mod.Structs {
			fmt.Printf("\t\t%d.- Struct: %s\n", j, strct.Name)

			if withAffs {
				for i, aff := range strct.GetAffordances() {
					fmt.Printf("\t\t * %d.- %s\n", i, aff.Description)
				}
			}

			for k, fld := range strct.Fields {
				fmt.Printf("\t\t\t%d.- Field: %s %s\n",
					k, fld.Name, fld.Typ.Name)
			}
			
			j++
		}

		if len(mod.Functions) > 0 {
			fmt.Println("\tFunctions")
		}

		j = 0
		for _, fn := range mod.Functions {

			inOuts := make(map[string]string)
			for _, in := range fn.Inputs {
				inOuts[in.Name] = in.Typ.Name
			}
			
			
			var inps bytes.Buffer
			for i, inp := range fn.Inputs {
				if i == len(fn.Inputs) - 1 {
					inps.WriteString(concat(inp.Name, " ", inp.Typ.Name))
				} else {
					inps.WriteString(concat(inp.Name, " ", inp.Typ.Name, ", "))
				}
			}

			var outs bytes.Buffer
			for i, out := range fn.Outputs {
				if i == len(fn.Outputs) - 1 {
					outs.WriteString(concat(out.Name, " ", out.Typ.Name))
				} else {
					outs.WriteString(concat(out.Name, " ", out.Typ.Name, ", "))
				}
			}

			// delete me
			//if fn.Name == "main" {
			fmt.Printf("\t\t%d.- Function: %s (%s) (%s)\n",
				j, fn.Name, inps.String(), outs.String())
			//}

			if withAffs {
				for i, aff := range fn.GetAffordances() {
					fmt.Printf("\t\t * %d.- %s\n", i, aff.Description)
				}
			}

			k := 0
			for _, expr := range fn.Expressions {
				//Arguments
				var args bytes.Buffer

				for i, arg := range expr.Arguments {
					//fmt.Println(string(*arg.Value))
					typ := ""
					if arg.Typ.Name == "ident" {
						if arg.Typ != nil &&
							inOuts[string(*arg.Value)] != "" {
							typ = inOuts[string(*arg.Value)]
						} else if arg.Value != nil { //&&
							// mod.Definitions[string(*arg.Value)] != nil &&
							// mod.Definitions[string(*arg.Value)].Typ.Name != ""
							//{

							//found := false
							var found *CXDefinition
							for _, def := range mod.Definitions {
								if def.Name == string(*arg.Value) {
									found = def
									break
								}
							}
							if found != nil && found.Typ.Name != "" {
								typ = found.Typ.Name
							}
							//typ = mod.Definitions[string(*arg.Value)].Typ.Name
						} else {
							typ = arg.Typ.Name
						}
					} else {
						typ = arg.Typ.Name
					}

					argName := string(*arg.Value)

					if arg.Typ.Name != "ident" {
						switch typ {
						case "str":
							argName = fmt.Sprintf("%#v", string(*arg.Value))
						case "bool":
							var val int32
							encoder.DeserializeRaw(*arg.Value, &val)
							if val == 0 {
								argName = "false"
							} else {
								argName = "true"
							}
						case "byte":
							argName = fmt.Sprintf("%#v", *arg.Value)
						case "i32":
							var val int32
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "i64":
							var val int64
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "f32":
							var val float32
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "f64":
							var val float64
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "[]byte":
							var val []byte
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "[]i32":
							var val []int32
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "[]i64":
							var val []int64
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "[]f32":
							var val []float32
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						case "[]f64":
							var val []float64
							encoder.DeserializeRaw(*arg.Value, &val)
							argName = fmt.Sprintf("%#v", val)
						default:
							argName = string(*arg.Value)
						}
					}

					if arg.Offset > -1 {
						offset := arg.Offset
						size := arg.Size
						var val []byte
						encoder.DeserializeRaw((*cxt.Heap)[offset:offset+size], &val)
						arg.Value = &val
					}

					if i == len(expr.Arguments) - 1 {
						args.WriteString(concat(argName, " ", typ))
					} else {
						args.WriteString(concat(argName, " ", typ, ", "))
					}
				}

				if len(expr.OutputNames) > 0 {
					var outNames bytes.Buffer
					for i, outName := range expr.OutputNames {
						// making outName shorter
						// if len(outName) > 15 {
						// 	outName = outName[:15]
						// }
						if i == len(expr.OutputNames) - 1 {
							outNames.WriteString(outName.Name)
						} else {
							outNames.WriteString(concat(outName.Name, ", "))
						}
					}
					
					fmt.Printf("\t\t\t%d.- Expression: %s = %s(%s)\n",
						k,
						outNames.String(),
						expr.Operator.Name,
						args.String())
				} else {
					fmt.Printf("\t\t\t%d.- Expression: %s(%s)\n",
						k,
						expr.Operator.Name,
						args.String())
				}

				

				if withAffs {
					for i, aff := range expr.GetAffordances() {
						fmt.Printf("\t\t\t * %d.- %s\n", i, aff.Description)
					}
				}
				
				k++
			}
			j++
		}
		i++
	}
}
