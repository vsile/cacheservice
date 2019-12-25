package main

import (
	"fmt"
    "log"
    "net/http"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
    "math/rand"
    "time"
    "strconv"
)

type cnf struct {
    URLs                []string    `yaml:"URLs"`
    MinTimeout          int         `yaml:"MinTimeout"`
    MaxTimeout          int         `yaml:"MaxTimeout"`
    NumberOfRequests    int         `yaml:"NumberOfRequests"`
}

type record struct {
    Id      bson.ObjectId   `_id`
    Url     string
    Status  string
}

func main() {
    //Подключаемся к mongoDB
	session, _ := mgo.Dial("localhost"); defer session.Close()
	c := session.DB("local").C("cache")

    //Считываем конфигурационный файл
    file, err := os.Open("config.yml")
    if err != nil { log.Fatal(err) }
    defer file.Close()

    body, _ := ioutil.ReadAll(file)
    config := cnf{}
    yaml.Unmarshal(body, &config)


    //Слушаем запросы от cacheConsumer.go
    http.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
        for i := 0; i < config.NumberOfRequests; i++ {
            //go func() {
                rand.Seed(time.Now().UnixNano())    //Рандомизируем генератор случайных чисел
                randUrlIndex := rand.Intn(len(config.URLs)) //Индекс случайного URL в массиве URLs
                createdAt := "createdAt_"+strconv.Itoa(randUrlIndex) //Ключевое имя для индекса, по которому будет определяться TTL

                randUrl := config.URLs[randUrlIndex]
                randExpire := rand.Intn(config.MaxTimeout-config.MinTimeout)+config.MinTimeout

                //Ищем запись в базе данных
                data := record{}
                err = c.Find(bson.M{"url": randUrl}).One(&data)
                if err != nil { //Если записи в базе данных нет
                    resp, err := http.Get(randUrl)    //Предполагаем, что этот get-запрос тарифицируется
                    if err != nil {
                        fmt.Fprintf(w, "Ошибка: "+err.Error())
                        return
                    }
                    defer resp.Body.Close()
                    status := resp.Status
                    msg := status+"\nДанные по URL-запросу ("+randUrl+")"
                    fmt.Println(msg)
                    fmt.Fprintf(w, msg)     //Отправляем пользователю данные (для примера, это статус запроса)

                    c.DropIndex(createdAt)   //Удаляем индекс с таким же именем в БД
                    index := mgo.Index{
                        Key: []string{createdAt},
                        //Unique: true,
                        //DropDups: true, //Удаляем документы, индексированные по ключу createdAt
                        //Background: true,
                        //Sparse: true,
                        ExpireAfter: time.Duration(randExpire)*time.Second, //TTL документов
                    }
                    c.EnsureIndex(index)    //Добавляем/обновляем индекс в коллекции

                    //Делаем новую запись в БД
                    c.Insert(bson.M{"url": randUrl, createdAt: time.Now(), "status": status})
                } else {    //Если запись в базе данных присутствует
                    msg := data.Status+"\nДанные из кэша ("+randUrl+")"
                    fmt.Println(msg)
                    fmt.Fprintf(w, msg) //Отправляем пользователю данные (для примера, это статус запроса)
                }
            //}()
        }
    })
    
	http.ListenAndServe(":4445", nil)
}

