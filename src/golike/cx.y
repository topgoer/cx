%{
	package main
	import (
		"strings"
		"bytes"
		"fmt"
		"os"
		"github.com/skycoin/skycoin/src/cipher/encoder"
		. "github.com/skycoin/cx/src/base"


		"github.com/mndrix/golog"
		"github.com/mndrix/golog/read"
		"github.com/mndrix/golog/term"
	)

	var cxt = MakeContext()
	var m = golog.NewInteractiveMachine()
	
	var lineNo int = 0
	var replMode bool = false
	var replTargetFn = ""
	var dStack bool = true
	var dProgram bool = false

	func warnf(format string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, format, args...)
		os.Stderr.Sync()
	}
%}

%union {
	i int
	i32 int32
	f32 float32
	i64 int64
	f64 float64
	tok string
	bool bool
	string string

	parameter *CXParameter
	parameters []*CXParameter

	argument *CXArgument
	arguments []*CXArgument

	expression *CXExpression
	expressions []*CXExpression

	field *CXField
	fields []*CXField

	name string
	names []string
}

%token  <i32>           INT BOOLEAN
%token  <f32>           FLOAT
%token  <tok>           FUNC OP LPAREN RPAREN LBRACE RBRACE IDENT
                        VAR COMMA COMMENT STRING PACKAGE IF ELSE WHILE TYPSTRUCT STRUCT
                        ASSIGN CASSIGN GTHAN LTHAN LTEQ GTEQ IMPORT
                        /* Types */
                        BOOL STR I32 I64 F32 F64 BYTE BOOLA BYTEA I32A I64A F32A F64A
                        /* Selectors */
                        SPACKAGE SSTRUCT SFUNC
                        /* Removers */
                        REM GLOBAL EXPR FIELD INPUT OUTPUT CLAUSES OBJECT OBJECTS
                        /* Stepping */
                        STEP PSTEP
                        /* Debugging */
                        DSTACK DPROGRAM DQUERY
                        /* Affordances */
                        AFF
                        /* Prolog */
                        CCLAUSES QUERY COBJECT COBJECTS

%type   <tok>           typeSpecifier
                        
%type   <parameter>     parameter
%type   <parameters>    parameters functionParameters
%type   <argument>      argument definitionAssignment
%type   <arguments>     arguments argumentsList
%type   <fields>        fields structFields
%type   <expression>    assignExpression
%type   <expressions>   elseStatement
%type   <name>          nonAssignExpression
%type   <bool>          selectorLines selectorExpressionsAndStatements selectorFields

%%

lines:
                /* empty */
        |       lines line
                {
			if replMode && dProgram {
				cxt.PrintProgram(false)
			}
                }
        ;

line:
                definitionDeclaration
        |       structDeclaration
        |       packageDeclaration
        |       importDeclaration
        |       functionDeclaration
        |       selector
        |       stepping
        |       debugging
        |       affordance
        |       removers
                // prolog
        |       prolog
        ;

// :clause OPERATOR ARGUMENT
// :clause IDENT IDENT IDENT
// :clause ROBOT TURNLEFT

prolog:
                COBJECTS
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				for i, object := range mod.Objects {
					fmt.Printf("%d.- %s\n", i, object)
				}
			}
                }
        |       CCLAUSES STRING
                {
			clauses := strings.TrimPrefix($2, "\"")
			clauses = strings.TrimSuffix(clauses, "\"")

			b := bytes.NewBufferString(clauses)
			m = m.Consult(b)
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.Clauses = clauses
			}
                }
        |       COBJECT IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.AddObject($2);
			}
                }
        |       QUERY STRING
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				query := strings.TrimPrefix($2, "\"")
				query = strings.TrimSuffix(query, "\"")
				mod.Query = query
			}
                }
        |       DQUERY STRING
                {
			query := strings.TrimPrefix($2, "\"")
			query = strings.TrimSuffix(query, "\"")

			goal, err := read.Term(query)
			if err == nil {
				variables := term.Variables(goal)
				answers := m.ProveAll(goal)

				yesNoAnswer := false
				if len(answers) == 0 {
					fmt.Println("no.")
					yesNoAnswer = true
				} else if variables.Size() == 0 {
					fmt.Println("yes.")
					yesNoAnswer = true
				}

				if !yesNoAnswer {
					for i, answer := range answers {
						lines := make([]string, 0)
						variables.ForEach(func(name string, variable interface{}) {
							v := variable.(*term.Variable)
							val := answer.Resolve_(v)
							line := fmt.Sprintf("%s = %s", name, val)
							lines = append(lines, line)
						})

						warnf("%s", strings.Join(lines, "\n"))
						if i == len(answers)-1 {
							fmt.Printf("\t.\n\n")
						} else {
							warnf("\t;\n")
						}
					}
				}
			} else {
				fmt.Println("Problem parsing the query.")
			}
                }
        ;

importDeclaration:
                IMPORT STRING
                {
			impName := strings.TrimPrefix($2, "\"")
			impName = strings.TrimSuffix(impName, "\"")
			if imp, err := cxt.GetModule(impName); err == nil {
				if mod, err := cxt.GetCurrentModule(); err == nil {
					mod.AddImport(imp)
				}
			}
                }
        ;

selectorLines:
                /* empty */
                {
			$$ = false
                }
        |       LBRACE lines RBRACE
                {
			$$ = true
                }
        ;

affordance:
                /* Function Affordances */
                AFF FUNC IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := cxt.GetFunction($3, mod.Name); err == nil {
					affs := fn.GetAffordances()
					for i, aff := range affs {
						fmt.Printf("(%d)\t%s\n", i, aff.Description)
					}
				}
			}
                }
        |       AFF FUNC IDENT LBRACE INT RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := cxt.GetFunction($3, mod.Name); err == nil {
					affs := fn.GetAffordances()
					affs[$5].ApplyAffordance()
				}
			}
                }
        |       AFF FUNC IDENT LBRACE STRING RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := cxt.GetFunction($3, mod.Name); err == nil {
					affs := fn.GetAffordances()
					filter := strings.TrimPrefix($5, "\"")
					filter = strings.TrimSuffix(filter, "\"")
					affs = FilterAffordances(affs, filter)
					for i, aff := range affs {
						fmt.Printf("(%d)\t%s\n", i, aff.Description)
					}
				}
			}
                }
        |       AFF FUNC IDENT LBRACE STRING INT RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := cxt.GetFunction($3, mod.Name); err == nil {
					affs := fn.GetAffordances()
					filter := strings.TrimPrefix($5, "\"")
					filter = strings.TrimSuffix(filter, "\"")
					affs = FilterAffordances(affs, filter)
					affs[$6].ApplyAffordance()
				}
			}
                }
                /* Module Affordances */
        |       AFF PACKAGE IDENT
                {
			if mod, err := cxt.GetModule($3); err == nil {
				affs := mod.GetAffordances()
				for i, aff := range affs {
					fmt.Printf("(%d)\t%s\n", i, aff.Description)
				}
			}
                }
        |       AFF PACKAGE IDENT LBRACE INT RBRACE
                {
			if mod, err := cxt.GetModule($3); err == nil {
				affs := mod.GetAffordances()
				affs[$5].ApplyAffordance()
			}
                }
        |       AFF PACKAGE IDENT LBRACE STRING RBRACE
                {
			if mod, err := cxt.GetModule($3); err == nil {
				affs := mod.GetAffordances()
				filter := strings.TrimPrefix($5, "\"")
				filter = strings.TrimSuffix(filter, "\"")
				affs = FilterAffordances(affs, filter)
				for i, aff := range affs {
					fmt.Printf("(%d)\t%s\n", i, aff.Description)
				}
			}
                }
        |       AFF PACKAGE IDENT LBRACE STRING INT RBRACE
                {
			if mod, err := cxt.GetModule($3); err == nil {
				affs := mod.GetAffordances()
				filter := strings.TrimPrefix($5, "\"")
				filter = strings.TrimSuffix(filter, "\"")
				affs = FilterAffordances(affs, filter)
				affs[$6].ApplyAffordance()
			}
                }
                /* Struct Affordances */
        |       AFF STRUCT IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if strct, err := cxt.GetStruct($3, mod.Name); err == nil {
					affs := strct.GetAffordances()
					for i, aff := range affs {
						fmt.Printf("(%d)\t%s\n", i, aff.Description)
					}
				}
			}
			
                }
        |       AFF STRUCT IDENT LBRACE INT RBRACE
                {

			if mod, err := cxt.GetCurrentModule(); err == nil {
				if strct, err := cxt.GetStruct($3, mod.Name); err == nil {
					affs := strct.GetAffordances()
					affs[$5].ApplyAffordance()
				}
			}
                }
        |       AFF STRUCT IDENT LBRACE STRING RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if strct, err := cxt.GetStruct($3, mod.Name); err == nil {
					affs := strct.GetAffordances()
					filter := strings.TrimPrefix($5, "\"")
					filter = strings.TrimSuffix(filter, "\"")
					affs = FilterAffordances(affs, filter)
					for i, aff := range affs {
						fmt.Printf("(%d)\t%s\n", i, aff.Description)
					}
				}
			}
                }
        |       AFF STRUCT IDENT LBRACE STRING INT RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if strct, err := cxt.GetStruct($3, mod.Name); err == nil {
					affs := strct.GetAffordances()
					filter := strings.TrimPrefix($5, "\"")
					filter = strings.TrimSuffix(filter, "\"")
					affs = FilterAffordances(affs, filter)
					affs[$6].ApplyAffordance()
				}
			}
                }
                /* Struct Affordances */
        |       AFF EXPR
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if expr, err := fn.GetCurrentExpression(); err == nil {
						affs := expr.GetAffordances()
						for i, aff := range affs {
							fmt.Printf("(%d)\t%s\n", i, aff.Description)
						}
					}
				}
			}
                }
        |       AFF EXPR LBRACE INT RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if expr, err := fn.GetCurrentExpression(); err == nil {
						affs := expr.GetAffordances()
						affs[$4].ApplyAffordance()
					}
				}
			}
                }
        |       AFF EXPR LBRACE STRING RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if expr, err := fn.GetCurrentExpression(); err == nil {
						affs := expr.GetAffordances()
						filter := strings.TrimPrefix($5, "\"")
						filter = strings.TrimSuffix(filter, "\"")
						affs = FilterAffordances(affs, filter)
						for i, aff := range affs {
							fmt.Printf("(%d)\t%s\n", i, aff.Description)
						}
					}
				}
			}
                }
        |       AFF EXPR LBRACE STRING INT RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if expr, err := fn.GetCurrentExpression(); err == nil {
						affs := expr.GetAffordances()
						filter := strings.TrimPrefix($4, "\"")
						filter = strings.TrimSuffix(filter, "\"")
						affs = FilterAffordances(affs, filter)
						affs[$5].ApplyAffordance()
					}
				}
			}
                }
        ;

stepping:       STEP INT
                {
			if $2 == 0 {
				// we do nothing
			} else {
				if $2 < 0 {
					nCalls := $2 * -1
					cxt.UnRun(int(nCalls))
				} else {
					cxt.Run(dStack, int($2))
				}
			}
                }
        // |       PSTEP INT
        //         {
			
        //         }
        ;

debugging:      DSTACK BOOLEAN
                {
			if $2 > 0 {
				dStack = true
                        } else {
				dStack = false
			}
                }
        |       DPROGRAM BOOLEAN
                {
			if $2 > 0 {
				dProgram = true
                        } else {
				dProgram = false
			}
                }
        ;

removers:       REM FUNC IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.RemoveFunction($3)
			}
                }
        |       REM PACKAGE IDENT
                {
			cxt.RemoveModule($3)
                }
        |       REM GLOBAL IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.RemoveDefinition($3)
			}
                }
        |       REM STRUCT IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.RemoveStruct($3)
			}
                }
        |       REM IMPORT STRING
                {
			impName := strings.TrimPrefix($3, "\"")
			impName = strings.TrimSuffix(impName, "\"")
			
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.RemoveImport(impName)
			}
                }
        |       REM EXPR INT FUNC IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.Context.GetFunction($5, mod.Name); err == nil {
					fn.RemoveExpression(int($3))
				}
			}
                }
        |       REM FIELD IDENT STRUCT IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if strct, err := cxt.GetStruct($5, mod.Name); err == nil {
					strct.RemoveField($3)
				}
				
			}
                }
        |       REM INPUT IDENT FUNC IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.Context.GetFunction($5, mod.Name); err == nil {
					fn.RemoveInput($3)
				}
			}
                }
        |       REM OUTPUT IDENT FUNC IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.Context.GetFunction($5, mod.Name); err == nil {
					fn.RemoveOutput($3)
				}
			}
                }
        |       REM CLAUSES
                {
			m = golog.NewInteractiveMachine()
                }
        |       REM OBJECT IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.RemoveObject($3)
			}
                }
        |       REM OBJECTS
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.RemoveObjects()
			}
                }
                // no, too complex. just wipe entire expression
        // |       REM ARG INT EXPR INT FUNC IDENT
        //         {
	// 		if mod, err := cxt.GetCurrentModule(); err == nil {
	// 			if fn, err := mod.Context.GetFunction($5, mod.Name); err == nil {
	// 				fn.RemoveExpression(int($3))
	// 			}
	// 		}
        //         }
        //        |REM OUTNAME
        ;

selectorExpressionsAndStatements:
                /* empty */
                {
			$$ = false
                }
        |       LBRACE expressionsAndStatements RBRACE
                {
			$$ = true
                }
        ;

selectorFields:
                /* empty */
                {
			$$ = false
                }
        |       LBRACE fields RBRACE
                {
			if strct, err := cxt.GetCurrentStruct(); err == nil {
				for _, fld := range $2 {
					fldFromParam := MakeField(fld.Name, fld.Typ)
					strct.AddField(fldFromParam)
				}
			}
			$$ = true
                }
        ;

selector:       SPACKAGE IDENT
                {
			var previousModule *CXModule
			if mod, err := cxt.GetCurrentModule(); err == nil {
				previousModule = mod
			} else {
				panic("A current module does not exist")
			}
			if _, err := cxt.SelectModule($2); err == nil {
				//fmt.Println(fmt.Sprintf("== Changed to package '%s' ==", mod.Name))
			} else {
				fmt.Println(err)
			}

			$<string>$ = previousModule.Name
                }
                selectorLines
                {
			if $<bool>4 {
				if _, err := cxt.SelectModule($<string>3); err == nil {
					//fmt.Println(fmt.Sprintf("== Changed to package '%s' ==", mod.Name))
				}
			}
                }
        |       SFUNC IDENT
                {
			var previousFunction *CXFunction
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				previousFunction = fn
			} else {
				panic("A current function does not exist")
			}
			if _, err := cxt.SelectFunction($2); err == nil {
				//fmt.Println(fmt.Sprintf("== Changed to function '%s' ==", fn.Name))
			} else {
				fmt.Println(err)
			}

			replTargetFn = $2
			$<string>$ = previousFunction.Name
                }
                selectorExpressionsAndStatements
                {
			if $<bool>4 {
				if _, err := cxt.SelectFunction($<string>3); err == nil {
					//fmt.Println(fmt.Sprintf("== Changed to function '%s' ==", fn.Name))
				}
			}
                }
        |       SSTRUCT IDENT
                {
			var previousStruct *CXStruct
			if fn, err := cxt.GetCurrentStruct(); err == nil {
				previousStruct = fn
			} else {
				panic("A current struct does not exist")
			}
			if _, err := cxt.SelectStruct($2); err == nil {
				//fmt.Println(fmt.Sprintf("== Changed to struct '%s' ==", fn.Name))
			} else {
				fmt.Println(err)
			}

			$<string>$ = previousStruct.Name
                }
                selectorFields
                {
			if $<bool>4 {
				if _, err := cxt.SelectStruct($<string>3); err == nil {
					//fmt.Println(fmt.Sprintf("== Changed to struct '%s' ==", strct.Name))
				}
			}
                }
        ;

assignOperator:
                ASSIGN
        |       CASSIGN
        ;

typeSpecifier:
                I32 {$$ = $1}
        |       I64 {$$ = $1}
        |       F32 {$$ = $1}
        |       F64 {$$ = $1}
        |       BOOL {$$ = $1}
        |       BYTE {$$ = $1}
        |       BOOLA {$$ = $1}
        |       BYTEA {$$ = $1}
        |       I32A {$$ = $1}
        |       I64A {$$ = $1}
        |       F32A {$$ = $1}
        |       F64A {$$ = $1}
        |       STR {$$ = $1}
        ;

packageDeclaration:
                PACKAGE IDENT
                {
			cxt.AddModule(MakeModule($2))
                }
                ;

definitionAssignment:
                /* empty */
                {
			$$ = nil
                }
        |       assignOperator argument
                {
			$$ = $2
                }
                ;

definitionDeclaration:
                VAR IDENT typeSpecifier definitionAssignment
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				var val *CXArgument;
				if $4 == nil {
					val = MakeArgument(MakeDefaultValue($3), MakeType($3))
				} else {
					switch $3 {
					case "byte":
						var ds int32
						encoder.DeserializeRaw(*$4.Value, &ds)
						//new := encoder.Serialize(byte(ds))
						new := []byte{byte(ds)}
						val = MakeArgument(&new, MakeType("byte"))
					case "i64":
						var ds int32
						encoder.DeserializeRaw(*$4.Value, &ds)
						new := encoder.Serialize(int64(ds))
						val = MakeArgument(&new, MakeType("i64"))
					case "f64":
						var ds float32
						encoder.DeserializeRaw(*$4.Value, &ds)
						new := encoder.Serialize(float64(ds))
						val = MakeArgument(&new, MakeType("f64"))
					default:
						val = $4
					}
				}
				
				mod.AddDefinition(MakeDefinition($2, val.Value, MakeType($3)))
			}
                }
        |       VAR IDENT IDENT
                {
			// we have to initialize all the fields
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if strct, err := cxt.GetStruct($3, mod.Name); err == nil {
					if flds, err := strct.GetFields(); err == nil {
						//_ = flds
						for _, fld := range flds {
							zeroVal := MakeDefaultValue(fld.Typ.Name)
							defName := fmt.Sprintf("%s.%s", $2, fld.Name)
							mod.AddDefinition(MakeDefinition(defName, zeroVal, fld.Typ))
						}
					}
				} else {
					panic(fmt.Sprintf("Type '%s' not defined", $3))
				}
			}
                }
        ;

fields:
                parameter
                {
			var flds []*CXField
                        flds = append(flds, MakeFieldFromParameter($1))
			$$ = flds
                }
        |       fields parameter
                {
			$1 = append($1, MakeFieldFromParameter($2))
			$$ = $1
                }
                ;

structFields:
                LBRACE fields RBRACE
                {
			$$ = $2
                }
        |       LBRACE RBRACE
                {
			$$ = nil
                }
        ;

structDeclaration:
                TYPSTRUCT IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				strct := MakeStruct($2)
				mod.AddStruct(strct)
			}
                }
                STRUCT structFields
                {
			if strct, err := cxt.GetCurrentStruct(); err == nil {
				for _, fld := range $5 {
					fldFromParam := MakeField(fld.Name, fld.Typ)
					strct.AddField(fldFromParam)
				}
			}
                }
        ;

functionParameters:
                LPAREN parameters RPAREN
                {
                    $$ = $2
                }
        |       LPAREN RPAREN
                {
                    $$ = nil
                }
        ;

functionDeclaration:
                FUNC IDENT functionParameters functionParameters
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				mod.AddFunction(MakeFunction($2))
				if fn, err := mod.GetCurrentFunction(); err == nil {
					for _, inp := range $3 {
						fn.AddInput(inp)
					}
					for _, out := range $4 {
						fn.AddOutput(out)
					}
				}
			}
                }
                functionStatements
        ;

parameter:
                IDENT typeSpecifier
                {
			$$ = MakeParameter($1, MakeType($2))
                }
        ;

parameters:
                parameter
                {
			var params []*CXParameter
                        params = append(params, $1)
                        $$ = params
                }
        |       parameters COMMA parameter
                {
			$1 = append($1, $3)
                        $$ = $1
                }
        ;

functionStatements:
                LBRACE expressionsAndStatements RBRACE
                /* { */
		/* 	$$ = $2 */
                /* } */
        |       LBRACE RBRACE
                /* { */
		/* 	$$ = nil */
                /* } */
        ;

expressionsAndStatements:
                nonAssignExpression
        |       assignExpression
        |       statement
        |       selector
        |       stepping
        |       debugging
        |       affordance
        |       removers
        |       prolog
        |       expressionsAndStatements nonAssignExpression
        |       expressionsAndStatements assignExpression
        |       expressionsAndStatements statement
        |       expressionsAndStatements selector
        |       expressionsAndStatements stepping
        |       expressionsAndStatements debugging
        |       expressionsAndStatements affordance
        |       expressionsAndStatements removers
        |       expressionsAndStatements prolog
        ;


assignExpression:
                VAR IDENT typeSpecifier definitionAssignment
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := cxt.GetCurrentFunction(); err == nil {
					if $4 == nil {
						if op, err := cxt.GetFunction("initDef", mod.Name); err == nil {
							expr := MakeExpression(op)
							fn.AddExpression(expr)
							expr.AddOutputName($2)
							typ := []byte($3)
							arg := MakeArgument(&typ, MakeType("str"))
							expr.AddArgument(arg)

							if strct, err := cxt.GetStruct($3, mod.Name); err == nil {
								for _, fld := range strct.Fields {
									expr := MakeExpression(op)
									fn.AddExpression(expr)
									expr.AddOutputName(fmt.Sprintf("%s.%s", $2, fld.Name))
									typ := []byte(fld.Typ.Name)
									arg := MakeArgument(&typ, MakeType("str"))
									expr.AddArgument(arg)
								}
							}
						}
					} else {
						switch $3 {
						case "byte":
							var ds int32
							encoder.DeserializeRaw(*$4.Value, &ds)
							new := []byte{byte(ds)}
							val := MakeArgument(&new, MakeType("byte"))
							
							if op, err := cxt.GetFunction("idByte", mod.Name); err == nil {
								expr := MakeExpression(op)
								fn.AddExpression(expr)
								expr.AddOutputName($2)
								expr.AddArgument(val)
							}
						case "i64":
							var ds int32
							encoder.DeserializeRaw(*$4.Value, &ds)
							new := encoder.Serialize(int64(ds))
							val := MakeArgument(&new, MakeType("i64"))

							if op, err := cxt.GetFunction("idI64", mod.Name); err == nil {
								expr := MakeExpression(op)
								fn.AddExpression(expr)
								expr.AddOutputName($2)
								expr.AddArgument(val)
							}
						case "f64":
							var ds float32
							encoder.DeserializeRaw(*$4.Value, &ds)
							new := encoder.Serialize(float64(ds))
							val := MakeArgument(&new, MakeType("f64"))

							if op, err := cxt.GetFunction("idF64", mod.Name); err == nil {
								expr := MakeExpression(op)
								fn.AddExpression(expr)
								expr.AddOutputName($2)
								expr.AddArgument(val)
							}
						default:
							val := $4
							var getFn string
							switch $3 {
							case "i32": getFn = "idI32"
							case "f32": getFn = "idF32"
							case "[]byte": getFn = "idByteA"
							case "[]i32": getFn = "idI32A"
							case "[]i64": getFn = "idI64A"
							case "[]f32": getFn = "idF32A"
							case "[]f64": getFn = "idF64A"
							}

							if op, err := cxt.GetFunction(getFn, mod.Name); err == nil {
								expr := MakeExpression(op)
								fn.AddExpression(expr)
								expr.AddOutputName($2)
								expr.AddArgument(val)
							}
						}
						
					}
				}
			}
                }
        |       argumentsList assignOperator IDENT arguments
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := cxt.GetCurrentFunction(); err == nil {
					if op, err := cxt.GetFunction($3, mod.Name); err == nil {
						expr := MakeExpression(op)

						for _, outNameInArg := range $1 {
							expr.AddOutputName(string(*outNameInArg.Value))
						}
						
						fn.AddExpression(expr)

						if expr, err := fn.GetCurrentExpression(); err == nil {
							for _, arg := range $4 {
								expr.AddArgument(arg)
							}
							$$ = expr
						}
						
					} else {
						panic(err)
					}
				}
				
			} else {
				panic(err)
			}
                }
        ;

nonAssignExpression:
                IDENT arguments
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := cxt.GetCurrentFunction(); err == nil {
					if op, err := cxt.GetFunction($1, mod.Name); err == nil {
						syntOutName := MakeGenSym("nonAssign")
						expr := MakeExpression(op)
						fn.AddExpression(expr)
						for _, arg := range $2 {
							expr.AddArgument(arg)
						}
						expr.AddOutputName(syntOutName)
						$$ = syntOutName
					} else {
						panic(fmt.Sprintf("Function '%s' not defined", $1))
					}
				}
			}
                }
        ;

statement:
                IF nonAssignExpression
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						expr := MakeExpression(goToFn)
						fn.AddExpression(expr)
					}
				}
			}
                }
                LBRACE
                {
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				$<i>$ = len(fn.Expressions)
			}
                }
                expressionsAndStatements RBRACE elseStatement
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
                                        goToExpr := fn.Expressions[$<i>5 - 1]

					var elseLines []byte
					if $<i>8 > 0 {
						elseLines = encoder.Serialize(int32(len(fn.Expressions) - $<i>5 - $<i>8 + 1))
					} else {
						elseLines = encoder.Serialize(int32(len(fn.Expressions) - $<i>5 + 1))
					}
					
					thenLines := encoder.Serialize(int32(1))

					predVal := []byte($2)

					goToExpr.AddArgument(MakeArgument(&predVal, MakeType("ident")))
					goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
					goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
				}
			}
                }
        |       IF IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						expr := MakeExpression(goToFn)
						fn.AddExpression(expr)
					}
				}
			}
                }
                LBRACE
                {
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				$<i>$ = len(fn.Expressions)
			}
                }
                expressionsAndStatements RBRACE elseStatement
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					goToExpr := fn.Expressions[$<i>5 - 1]
					var elseLines []byte
					if $<i>8 > 0 {
						elseLines = encoder.Serialize(int32(len(fn.Expressions) - $<i>5 - $<i>8 + 1))
					} else {
						elseLines = encoder.Serialize(int32(len(fn.Expressions) - $<i>5 + 1))
					}
					//elseLines := encoder.Serialize(int32(len(fn.Expressions) - $<i>5 - 3))
					thenLines := encoder.Serialize(int32(1))

					predVal := []byte($2)

					goToExpr.AddArgument(MakeArgument(&predVal, MakeType("ident")))
					goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
					goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
				}
			}
                }
        |       IF BOOLEAN
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						expr := MakeExpression(goToFn)
						fn.AddExpression(expr)
					}
				}
			}
                }
                LBRACE
                {
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				$<i>$ = len(fn.Expressions)
			}
                }
                expressionsAndStatements RBRACE elseStatement
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					goToExpr := fn.Expressions[$<i>5 - 1]
					var elseLines []byte
					if $<i>8 > 0 {
						elseLines = encoder.Serialize(int32(len(fn.Expressions) - $<i>5 - $<i>8 + 1))
					} else {
						elseLines = encoder.Serialize(int32(len(fn.Expressions) - $<i>5 + 1))
					}
					//elseLines := encoder.Serialize(int32(len(fn.Expressions) - $<i>5 - 3))
					thenLines := encoder.Serialize(int32(1))

					predVal := encoder.Serialize($2)
					
					goToExpr.AddArgument(MakeArgument(&predVal, MakeType("bool")))
					goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
					goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
				}
			}
                }
        |       WHILE nonAssignExpression
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						expr := MakeExpression(goToFn)
						fn.AddExpression(expr)
					}
				}
			}
                }
                LBRACE
                {
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				$<i>$ = len(fn.Expressions)
			}
                }
                expressionsAndStatements RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					goToExpr := fn.Expressions[$<i>5 - 1]
					elseLines := encoder.Serialize(int32(len(fn.Expressions) - $<i>5 + 2))
					thenLines := encoder.Serialize(int32(1))
					
					predVal := []byte($2)
					
					goToExpr.AddArgument(MakeArgument(&predVal, MakeType("ident")))
					goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
					goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
					
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						goToExpr := MakeExpression(goToFn)
						fn.AddExpression(goToExpr)

						elseLines := encoder.Serialize(int32(0))
						thenLines := encoder.Serialize(int32(-len(fn.Expressions) + $<i>5 - 1))

						alwaysTrue := encoder.Serialize(int32(1))

						goToExpr.AddArgument(MakeArgument(&alwaysTrue, MakeType("bool")))
						goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
						goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
					}
					
				}
			}
                }
        |       WHILE IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						expr := MakeExpression(goToFn)
						fn.AddExpression(expr)
					}
				}
			}
                }
                LBRACE
                {
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				$<i>$ = len(fn.Expressions)
			}
                }
                expressionsAndStatements RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					goToExpr := fn.Expressions[$<i>5 - 1]

					elseLines := encoder.Serialize(int32(len(fn.Expressions) - $<i>5 + 2))
					thenLines := encoder.Serialize(int32(1))
					
					predVal := []byte($2)
					
					goToExpr.AddArgument(MakeArgument(&predVal, MakeType("ident")))
					goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
					goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
					
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						goToExpr := MakeExpression(goToFn)
						fn.AddExpression(goToExpr)

						elseLines := encoder.Serialize(int32(0))
						thenLines := encoder.Serialize(int32(-len(fn.Expressions) + $<i>5))

						alwaysTrue := []byte{1}

						goToExpr.AddArgument(MakeArgument(&alwaysTrue, MakeType("byte")))
						goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
						goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
					}
					
				}
			}
                }
        |       WHILE BOOLEAN
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						expr := MakeExpression(goToFn)
						fn.AddExpression(expr)
					}
				}
			}
                }
                LBRACE
                {
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				$<i>$ = len(fn.Expressions)
			}
                }
                expressionsAndStatements RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					goToExpr := fn.Expressions[$<i>5 - 1]
					
					elseLines := encoder.Serialize(int32(len(fn.Expressions) - $<i>5 + 2))
					thenLines := encoder.Serialize(int32(1))
					
					var predVal []byte
					if $2 == int32(1) {
						predVal = []byte{1}
					} else {
						predVal = []byte{0}
					}

					goToExpr.AddArgument(MakeArgument(&predVal, MakeType("byte")))
					goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
					goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
					
					if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
						goToExpr := MakeExpression(goToFn)
						fn.AddExpression(goToExpr)

						elseLines := encoder.Serialize(int32(0))
						thenLines := encoder.Serialize(int32(-len(fn.Expressions) + $<i>5))

						alwaysTrue := []byte{1}

						goToExpr.AddArgument(MakeArgument(&alwaysTrue, MakeType("byte")))
						goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
						goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))
					}
					
				}
			}
                }
        |       VAR IDENT IDENT
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if strct, err := cxt.GetStruct($3, mod.Name); err == nil {
					if flds, err := strct.GetFields(); err == nil {
						//_ = flds
						for _, fld := range flds {
							zeroVal := MakeDefaultValue(fld.Typ.Name)
							localName := fmt.Sprintf("%s.%s", $2, fld.Name)
							if fn, err := cxt.GetCurrentFunction(); err == nil {
								if op, err := cxt.GetFunction(MakeIdentityOpName(fld.Typ.Name), mod.Name); err == nil {
									expr := MakeExpression(op)
									fn.AddExpression(expr)
									
									expr.AddArgument(MakeArgument(zeroVal, fld.Typ))
									expr.AddOutputName(localName)
								}
							}
						}
					}
				} else {
					panic(fmt.Sprintf("Type '%s' not defined", $3))
				}
			}
                }
        ;

elseStatement:
                /* empty */
                {
                    $<i>$ = 0
                }
        |       ELSE
                {
                    if mod, err := cxt.GetCurrentModule(); err == nil {
                        if fn, err := mod.GetCurrentFunction(); err == nil {
                            if goToFn, err := cxt.GetFunction("goTo", mod.Name); err == nil {
                            expr := MakeExpression(goToFn)
                            fn.AddExpression(expr)
                            }
                        }
                    }
                }
                LBRACE
                {
			if fn, err := cxt.GetCurrentFunction(); err == nil {
				$<i>$ = len(fn.Expressions)
			}
                }
                expressionsAndStatements RBRACE
                {
			if mod, err := cxt.GetCurrentModule(); err == nil {
				if fn, err := mod.GetCurrentFunction(); err == nil {
					goToExpr := fn.Expressions[$<i>4 - 1]
					
					elseLines := encoder.Serialize(int32(0))
					thenLines := encoder.Serialize(int32(len(fn.Expressions) - $<i>4 + 1))

					alwaysTrue := encoder.Serialize(int32(1))

					goToExpr.AddArgument(MakeArgument(&alwaysTrue, MakeType("bool")))
					goToExpr.AddArgument(MakeArgument(&thenLines, MakeType("i32")))
					goToExpr.AddArgument(MakeArgument(&elseLines, MakeType("i32")))

					$<i>$ = len(fn.Expressions) - $<i>4
				}
			}
                }
                ;

expressions:
                nonAssignExpression
        |       assignExpression
        |       expressions nonAssignExpression
        |       expressions assignExpression
        ;

argument:
                INT
                {
			val := encoder.Serialize($1)
                        $$ = MakeArgument(&val, MakeType("i32"))
                }
        |       FLOAT
                {
			
			val := encoder.Serialize($1)
			$$ = MakeArgument(&val, MakeType("f32"))
                }
        |       STRING
                {
			str := strings.TrimPrefix($1, "\"")
                        str = strings.TrimSuffix(str, "\"")
                        
                        val := []byte(str)
                        $$ = MakeArgument(&val, MakeType("str"))
                }
        |       IDENT
                {
			val := []byte($1)
                        $$ = MakeArgument(&val, MakeType("ident"))
                }
        |       typeSpecifier LBRACE argumentsList RBRACE
                {
			switch $1 {
			case "[]byte":
                                vals := make([]byte, len($3))
				for i, arg := range $3 {
					var val int32
					encoder.DeserializeRaw(*arg.Value, &val)
					vals[i] = byte(val)
				}
				sVal := encoder.Serialize(vals)
				$$ = MakeArgument(&sVal, MakeType("[]byte"))
			case "[]i32":
                                vals := make([]int32, len($3))
				for i, arg := range $3 {
					var val int32
					encoder.DeserializeRaw(*arg.Value, &val)
					vals[i] = val
				}
				sVal := encoder.Serialize(vals)
				$$ = MakeArgument(&sVal, MakeType("[]i32"))
			case "[]i64":
                                vals := make([]int64, len($3))
				for i, arg := range $3 {
					var val int32
					encoder.DeserializeRaw(*arg.Value, &val)
					vals[i] = int64(val)
				}
				sVal := encoder.Serialize(vals)
				$$ = MakeArgument(&sVal, MakeType("[]i64"))
			case "[]f32":
                                vals := make([]float32, len($3))
				for i, arg := range $3 {
					var val float32
					encoder.DeserializeRaw(*arg.Value, &val)
					vals[i] = val
				}
				sVal := encoder.Serialize(vals)
				$$ = MakeArgument(&sVal, MakeType("[]f32"))
			case "[]f64":
                                vals := make([]float64, len($3))
				for i, arg := range $3 {
					var val float32
					encoder.DeserializeRaw(*arg.Value, &val)
					vals[i] = float64(val)
				}
				sVal := encoder.Serialize(vals)
				$$ = MakeArgument(&sVal, MakeType("[]f64"))
			}
                }
        ;

arguments:
                LPAREN argumentsList RPAREN
                {
                    $$ = $2
                }
        |       LPAREN RPAREN
                {
                    $$ = nil
                }
        ;

argumentsList:
                argument
                {
			var args []*CXArgument
			args = append(args, $1)
			$$ = args
                }
        |       nonAssignExpression
                {
			var args []*CXArgument
			identName := []byte($1)
			arg := MakeArgument(&identName, MakeType("ident"))
			args = append(args, arg)
			$$ = args
                }
        |       argumentsList COMMA argument
                {
			$1 = append($1, $3)
			$$ = $1
                }
        |       argumentsList COMMA nonAssignExpression
                {
			identName := []byte($3)
			arg := MakeArgument(&identName, MakeType("ident"))
			$1 = append($1, arg)
			$$ = $1
                }
        ;

%%

