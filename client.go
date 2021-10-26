package main

import (
    "fmt"
    "net/rpc"
    "os"
    "bufio"
)

const (
    AGREGAR = iota + 1
    PROMEDIO_ALUMNO
    PROMEDIO_GENERAL
    PROMEDIO_MATERIA
    SALIR = 0
)

type Args struct {
    ID uint64
    Nombre, Materia string
    Cal float64
}

func client() {
    scanner := bufio.NewScanner(os.Stdin)
    op := -1
    c, err := rpc.Dial("tcp", ":9999")
    if err != nil {
        fmt.Println(err)
        return
    }
    for op != SALIR {
        fmt.Print("\nSeleccione una opción:\n")
        fmt.Print(AGREGAR, ") Agregar calificación de una materia\n")
        fmt.Print(PROMEDIO_ALUMNO, ") Mostrar el promedio de un alumno\n")
        fmt.Print(PROMEDIO_GENERAL, ") Mostrar el promedio general\n")
        fmt.Print(PROMEDIO_MATERIA, ") Mostrar el promedio de una materia\n")
        fmt.Print(SALIR, ") Salir\n>> ")
        fmt.Scanln(&op)
        switch op {
        case AGREGAR:
            var cal float64
            var tmp int
            fmt.Print("Nombre: ")
            scanner.Scan()
            nom := scanner.Text()
            fmt.Print("Materia: ")
            scanner.Scan()
            mat := scanner.Text()
            fmt.Print("Calificación: ")
            fmt.Scanln(&cal)
            err = c.Call("Server.AddGrade", Args{Nombre: nom, Materia: mat, Cal: cal}, &tmp)
            if err != nil { fmt.Println(err) }
            break
        case PROMEDIO_ALUMNO:
            var res float64
            var in uint64
            fmt.Print("Nombre: ")
            fmt.Scanln(&in)
            err = c.Call("Server.StudentMean", Args{ID: in}, &res)
            if err != nil {
                fmt.Println(err)
            } else {
                fmt.Printf("Promedio del alumno %d es: %f\n", in, res)
            }
            break
        case PROMEDIO_GENERAL:
            var res float64
            err = c.Call("Server.GeneralMean", Args{}, &res)
            if err != nil {
                fmt.Println(err)
            } else {
                fmt.Printf("El promedio general de los alumnos es: %f\n", res)
            }
            break
        case PROMEDIO_MATERIA:
            var res float64
            var in uint64
            fmt.Print("Materia: ")
            fmt.Scanln(&in)
            err = c.Call("Server.ClassMean", Args{ID: in}, &res)
            if err != nil {
                fmt.Println(err)
            } else {
                fmt.Printf("Promedio de la materia %d es: %f\n", in, res)
            }
            break
        case SALIR:
            return
        default: fmt.Println("Opción no válida, vuelva a intentarlo")
        }
    }
}

func main() {
    client()
}
