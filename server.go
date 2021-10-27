package main

import (
    "errors"
    "fmt"
    "encoding/json"
    "net"
    "net/rpc"
    "net/http"
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

var server *Server

type Args struct {
    ID uint64
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
    Add(args)
    return nil
}

func (t *Server) mean(tp string, id uint64) float64 {
    var res float64
    var n float64
    var m map[uint64]Node
    if tp == "student" {
        m = t.Alumnos[id].Value
    } else if tp == "class" {
        m = t.Materias[id].Value
    }
    for _, v := range m {
        res += v.Value
        n++
    }
    res /= n
    return res
}

func (t *Server) generalMean() float64 {
    var res float64
    var n float64
    for k := range t.Alumnos {
        res += t.mean("student", k)
        n++
    }
    res /= n
    return res
}

func (t *Server) StudentMean(args Args, reply *float64) error {
    if _, err := t.Alumnos[args.ID]; !err {
        return errors.New("El usuario " + args.Nombre + " no fue registrado con anterioridad")
    }
    (*reply) = t.mean("student", args.ID)
    return nil
}

func (t *Server) GeneralMean(args Args, reply *float64) error {
    if len(t.Alumnos) == 0 {
        return errors.New("No hay alumnos registrados")
    }
    (*reply) = t.generalMean()
    return nil
}

func (t *Server) ClassMean(args Args, reply *float64) error  {
    if _, err := t.Materias[args.ID]; !err {
        return errors.New("La materia " + args.Materia + " no fue registrada con anterioridad")
    }
    var res float64
    var n float64
    for _, v := range t.Materias[args.ID].Value {
        res += v.Value
        n++
    }
    res /= n
    (*reply) = res
    return nil
}

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

func Add(args Args) {
    fmt.Println()
    a := exists((*server).Alumnos, args.Nombre)
    m := exists((*server).Materias, args.Materia)
    if a == uint64(len((*server).Alumnos)) {
        (*server).Alumnos[a] = InnerMap{ Name: args.Nombre, Value: make(map[uint64]Node) }
        (*server).Alumnos[a].Value[m] = Node{ Name: args.Materia, Value: args.Cal }
        fmt.Printf("[Nuevo alumno añadido: %s]\n", args.Nombre)
    } else {
        (*server).Alumnos[a].Value[m] = Node{ Name: args.Materia, Value: args.Cal }
    }
    if m == uint64(len((*server).Materias)) {
        (*server).Materias[m] = InnerMap{ Name: args.Materia, Value: make(map[uint64]Node) }
        (*server).Materias[m].Value[a] = Node{ Name: args.Nombre, Value: args.Cal }
        fmt.Printf("[Nueva materia añadida: %s]\n", args.Materia)
    } else {
        (*server).Materias[m].Value[a] = Node{ Name: args.Nombre, Value: args.Cal }
    }
    printData("Alumnos: ", (*server).Alumnos)
    printData("Materias: ", (*server).Materias)
    fmt.Println("-----------------------------------------")
}

// TODO: Agregar alumno, materia y calificación (POST)
func add(res http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "POST":
        var args Args
        err := json.NewDecoder(req.Body).Decode(&args)
        if err != nil {
            http.Error(res, err.Error(), http.StatusBadRequest)
            return
        }
        Add(args)
        res_json := []byte(`{"code": "ok"}`)
        res.Header().Set("Content-Type", "application/json")
        res.Write(res_json)
    }
}

// TODO: Devolver al cliente todos los alumnos junto a su lista de materias y calificación
func get(res http.ResponseWriter, req *http.Request) {
    
}

// TODO: Devolver al cliente las materias (con calificación) de un alumno por id (GET/{id})
func getID(res http.ResponseWriter, req *http.Request) {
    
}

// TODO: Eliminar por id un alumno (DELETE/{id})
func del(res http.ResponseWriter, req *http.Request) {
    
}

// TODO: Modificar la calificación de un alumno (PUT/JSON)
func update(res http.ResponseWriter, req *http.Request) {
    
}

func main() {
    s := new(Server)
    s.Alumnos = make(map[uint64]InnerMap)
    s.Materias = make(map[uint64]InnerMap)
    go handleRpc(s)
    // Pointer used for a singleton style
    server = s
    // Peticiones HTTP
    http.HandleFunc("/add", add)
    http.HandleFunc("/data", get)
    http.HandleFunc("/data/{id}", getID)
    http.HandleFunc("/delete/{id}", del)
    http.HandleFunc("/modify", update)
    http.ListenAndServe(":9000", nil)
}
