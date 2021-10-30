package main

import (
    "errors"
    "fmt"
    "strings"
    "strconv"
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
    matID, alID uint64
}

type ModJson struct {
    Alumno, Materia uint64
    Cal float64
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

func exists(m map[uint64]InnerMap, n string, d *uint64) (uint64, bool) {
    for k, v := range m {
        if v.Name == n {
            return k, true
        }
    }
    return (*d), false
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

func AddStudent(args Args) {
    a, af := exists((*server).Alumnos, args.Nombre, &(*server).alID)
    m, _ := exists((*server).Materias, args.Materia, &(*server).matID)
    if !af {
        (*server).Alumnos[(*server).alID] = InnerMap{
            Name: args.Nombre,
            Value: make(map[uint64]Node),
        }
        (*server).Alumnos[(*server).alID].Value[m] = Node{
            Name: args.Materia,
            Value: args.Cal,
        }
        (*server).alID++
        fmt.Printf("[Nuevo alumno añadido: %s]\n", args.Nombre)
    } else {
        (*server).Alumnos[a].Value[m] = Node{
            Name: args.Materia,
            Value: args.Cal,
        }
    }
}

func AddGrade(args Args) {
    a, _ := exists((*server).Alumnos, args.Nombre, &(*server).alID)
    m, mf := exists((*server).Materias, args.Materia, &(*server).matID)
    if !mf {
        (*server).Materias[(*server).matID] = InnerMap{
            Name: args.Materia,
            Value: make(map[uint64]Node),
        }
        (*server).Materias[(*server).matID].Value[a] = Node{
            Name: args.Nombre,
            Value: args.Cal,
        }
        (*server).matID++
        fmt.Printf("[Nueva materia añadida: %s]\n", args.Materia)
    } else {
        (*server).Materias[m].Value[a] = Node{
            Name: args.Nombre,
            Value: args.Cal,
        }
    }
}

// Añade un usuario con su respectiva calificación y materia a los dos maps del server
func Add(args Args) {
    fmt.Println()
    AddStudent(args)
    AddGrade(args)
    printData("Alumnos: ", (*server).Alumnos)
    printData("Materias: ", (*server).Materias)
    fmt.Println("-----------------------------------------")
}

func Get(idStr string, res *http.ResponseWriter) []byte {
    var res_json []byte
    var err error
    // Devolver al cliente las materias (con calificación) de un alumno por id (GET/{id})
    if idStr != "/data" {
        id, _ := strconv.ParseUint(idStr, 10, 64)
        if _, exists := (*server).Alumnos[id]; exists {
            res_json, err = json.MarshalIndent((*server).Alumnos[id], "", "    ")
        } else {
            http.Error(*res, "El alumno proporcionado no existe", http.StatusNotFound)
        }
    // Devolver al cliente todos los alumnos junto a su lista de materias y calificación
    } else {
        // Si aún no hay alumnos registrados
        if len((*server).Alumnos) == 0 {
            res_json = []byte(`{"code": "empty"}`)
        } else {
            res_json, err = json.MarshalIndent((*server).Alumnos, "", "    ")
        }
    }
    if err != nil {
        http.Error(*res, err.Error(), http.StatusInternalServerError)
        res_json = []byte(`{"code": "error"}`)
    }
    return res_json
}

func Delete(id uint64, res *http.ResponseWriter) []byte {
    var res_json []byte
    if _, exists := (*server).Alumnos[id]; !exists {
        http.Error(*res, "El alumno proporcionado no existe", http.StatusBadRequest)
        return res_json
    }
    for k, v := range (*server).Materias {
        if _, exists := v.Value[id]; exists {
            delete((*server).Materias[k].Value, id)
        }
    }
    name := (*server).Alumnos[id].Name
    delete((*server).Alumnos, id)
    fmt.Printf("[El alumno %s ha sido eliminado exitosamente]\n", name)
    res_json = []byte(`{"code": "ok"}`)
    return res_json
}

func Put(a ModJson, res *http.ResponseWriter) []byte {
    var res_json []byte
    if _, exists := (*server).Alumnos[a.Alumno]; !exists {
        http.Error(*res, "El alumno proporcionado no existe", http.StatusBadRequest)
        return res_json
    } else if _, exists := (*server).Materias[a.Materia]; !exists {
        http.Error(*res, "La materia proporcionada no existe", http.StatusBadRequest)
        return res_json
    }
    node := (*server).Alumnos[a.Alumno].Value[a.Materia]
    node.Value = a.Cal
    (*server).Alumnos[a.Alumno].Value[a.Materia] = node

    node = (*server).Materias[a.Materia].Value[a.Alumno]
    node.Value = a.Cal
    (*server).Materias[a.Materia].Value[a.Alumno] = node
    res_json = []byte(`{"code": "ok"}`)
    return res_json
}

func CrudHandler(res http.ResponseWriter, req *http.Request) {
    var res_json []byte
    switch req.Method {
    // Agregar alumno, materia y calificación
    case "POST":
        var args Args
        err := json.NewDecoder(req.Body).Decode(&args)
        if err != nil {
            http.Error(res, err.Error(), http.StatusBadRequest)
            return
        }
        Add(args)
        res_json = []byte(`{"code": "ok"}`)
        res.Header().Set("Content-Type", "application/json")
        res.Write(res_json)
    // Devuelve alumnos
    case "GET":
        idStr := strings.TrimPrefix(req.URL.Path, "/data/")
        res_json := Get(idStr, &res)
        res.Header().Set("Content-Type", "application/json")
        res.Write(res_json)
    // Eliminar por id un alumno (DELETE/{id})
    case "DELETE":
        id, err := strconv.ParseUint(strings.TrimPrefix(req.URL.Path, "/data/"), 10, 64)
        if err != nil {
            http.Error(res, "El alumno proporcionado no existe", http.StatusNotFound)
            return
        }
        res_json = Delete(id, &res)
        res.Header().Set("Content-Type", "application/json")
        res.Write(res_json)
    // TODO: Modificar la calificación de un alumno (PUT/JSON)
    case "PUT":
        var a ModJson
        err := json.NewDecoder(req.Body).Decode(&a)
        if err != nil {
            http.Error(res, err.Error(), http.StatusBadRequest)
            return
        }
        res_json = Put(a, &res)
        res.Header().Set("Content-Type", "application/json")
        res.Write(res_json)
    }
}

func main() {
    s := new(Server)
    s.Alumnos = make(map[uint64]InnerMap)
    s.Materias = make(map[uint64]InnerMap)
    fmt.Println("Iniciando server RPC...")
    go handleRpc(s)
    // Puntero usado para diseño singleton
    server = s
    // Peticiones HTTP
    http.HandleFunc("/add", CrudHandler) // POST (añadir alumno)
    http.HandleFunc("/data", CrudHandler) // GET (Obtener todos los alumnos)
    http.HandleFunc("/data/", CrudHandler) // GET (Obtener un alumno por id)
    http.HandleFunc("/delete/", CrudHandler) // DELETE (Eliminar un alumno por id)
    http.HandleFunc("/modify", CrudHandler) // PUT (Modificar un alumno por JSON)
    fmt.Println("Iniciando server HTTP...")
    http.ListenAndServe(":9000", nil)
}
