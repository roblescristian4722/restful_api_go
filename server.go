package main

import (
    // "errors"
    "fmt"
    "net"
    "net/rpc"
)

type Node struct {
    Name string
    Value float64
}

type InnerMap struct {
    Name string
    Value map[uint64] Node
}

type Server struct {
    Materias, Alumnos map[uint64] InnerMap
}

type Args struct {
    Nombre, Materia string
    Cal float64
}

func printData(title string, m map[uint64]InnerMap) {
    fmt.Println(title)
    for k, v := range m {
        fmt.Printf("    %d) %s:\n", k, v.Name)
        for ki, vi := range v.Value {
            fmt.Printf("        %d) %s: %f\n", ki, vi.Name, vi.Value)
        }
    }
}

func exists(m map[uint64]InnerMap, n string) uint64 {
    for k, v := range m {
        if v.Name == n {
            return k
        }
    }
    return uint64(len(m))
}

func (t *Server) AddGrade(args Args, reply *int) error {
    fmt.Println()
    a := exists(t.Alumnos, args.Nombre)
    m := exists(t.Materias, args.Materia)
    if a == uint64(len(t.Alumnos)) {
        t.Alumnos[a] = InnerMap{ Name: args.Nombre, Value: make(map[uint64]Node) }
        t.Alumnos[a].Value[m] = Node{ Name: args.Materia, Value: args.Cal }
        fmt.Printf("[Nuevo alumno añadido: %s]\n", args.Nombre)
    } else {
        t.Alumnos[a].Value[m] = Node{ Name: args.Materia, Value: args.Cal }
    }
    if m == uint64(len(t.Materias)) {
        t.Materias[m] = InnerMap{ Name: args.Materia, Value: make(map[uint64]Node) }
        t.Materias[m].Value[a] = Node{ Name: args.Nombre, Value: args.Cal }
        fmt.Printf("[Nueva materia añadida: %s]\n", args.Materia)
    } else {
        t.Materias[m].Value[a] = Node{ Name: args.Nombre, Value: args.Cal }
    }
    printData("Alumnos: ", t.Alumnos)
    printData("Materias: ", t.Materias)
    fmt.Println("-----------------------------------------")
    return nil
}

// func (t *Server) AddGrade(args Args, reply *int) error {
//     fmt.Println()
//     if _, err := t.Alumnos[args.Nombre]; err {
//         t.Alumnos[args.Nombre][args.Materia] = args.Cal
//     } else {
//         fmt.Printf("[Nuevo alumno añadido: %s]\n", args.Nombre)
//         m := make(map[string] float64)
//         m[args.Materia] = args.Cal
//         t.Alumnos[args.Nombre] = m
//     }
//     if _, err := t.Materias[args.Materia]; err {
//         t.Materias[args.Materia][args.Nombre] = args.Cal
//     } else {
//         fmt.Printf("[Nueva materia añadida: %s]\n", args.Materia)
//         n := make( map[string] float64 )
//         n[args.Nombre] = args.Cal
//         t.Materias[args.Materia] = n
//     }
//     printData("Alumnos: ", t.Alumnos)
//     printData("Materias: ", t.Materias)
//     fmt.Println("-----------------------------------------")
//     return nil
// }

// func (t *Server) studentMean(name string) float64 {
//     var res float64
//     var n float64
//     for _, v := range t.Alumnos[name] {
//         res += v
//         n++
//     }
//     res /= n
//     return res
// }

// func (t *Server) StudentMean(args Args, reply *float64) error {
//     if _, err := t.Alumnos[args.Nombre]; !err {
//         return errors.New("El usuario " + args.Nombre + " no fue registrado con anterioridad")
//     }
//     (*reply) = t.studentMean(args.Nombre)
//     return nil
// }

// func (t *Server) GeneralMean(args Args, reply *float64) error {
//     if len(t.Alumnos) == 0 {
//         return errors.New("No hay alumnos registrados")
//     }
//     var res float64
//     var n float64
//     for k, _ := range t.Alumnos {
//         res += t.studentMean(k)
//         n++
//     }
//     res /= n
//     (*reply) = res
//     return nil
// }

// func (t *Server) ClassMean(args Args, reply *float64) error  {
//     if _, err := t.Materias[args.Materia]; !err {
//         return errors.New("La materia " + args.Materia + " no fue registrada con anterioridad")
//     }
//     var res float64
//     var n float64
//     for _, v := range t.Materias[args.Materia] {
//         res += v
//         n++
//     }
//     res /= n
//     (*reply) = res
//     return nil
// }

func handleRpc(s *Server) {
    rpc.Register(s)
    rpc.HandleHTTP()
    ln, err := net.Listen("tcp", ":9999")
    if err != nil {
        fmt.Println(err)
        return
    }
    for {
        c, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        go rpc.ServeConn(c)
    }
}

func main() {
    arith := new(Server)
    arith.Alumnos = make(map[uint64]InnerMap)
    arith.Materias = make(map[uint64]InnerMap)
    go handleRpc(arith)
    fmt.Scanln()
}
