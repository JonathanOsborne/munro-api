package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Munro struct {
	Altitude    int64    `json:"altitude,omitempty"`
	Description string   `json:"description,omitempty"`
	Name        string   `json:"name,omitempty"`
	Climbers    int64    `json:"climbers,omitempty"`
	Rating      float64  `json:"rating,omitempty"`
	Region      string   `json:"region,omitempty"`
	Routes      []string `json:"routes,omitempty"`
	Link        string   `json:"link,omitempty"`
	ID          string   `json:"id,omitempty"`
}

type GetMunroResponse struct {
	Munro  Munro  `json:"munro"`
	Index  int    `json:"index"`
	Prev   string `json:"prev"`
	Next   string `json:"next"`
	Random string `json:"random"`
}

func (c *Controller) SetMunrosData() {
	jsonFile, err := os.Open("munro_data.json")

	if err != nil {
		fmt.Println(err)
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("err")
	}

	var Munros map[string]Munro

	json.Unmarshal(byteValue, &Munros)

	c.Munros = Munros
	c.MongoClient.SetCollection("munros", "munros")

}

func (con *Controller) GetMunro(c *gin.Context) {
	name := c.Param("name")
	var munro Munro
	filter := bson.D{primitive.E{Key: "id", Value: name}}
	err := con.MongoClient.Collection.FindOne(con.MongoClient.Ctx, filter).Decode(&munro)

	if err == mongo.ErrNoDocuments {
		fmt.Println("record does not exist")
	} else if err != nil {
		log.Fatal(err)
	}

	var munroNames []string
	for k := range con.Munros {
		munroNames = append(munroNames, k)
	}

	sort.Strings(munroNames)
	index := -1
	for k, v := range munroNames {
		if name == v {
			index = k
		}
	}

	rand.Seed(time.Now().UnixNano())

	next := munroNames[index+1]
	prev := munroNames[index-1]
	random := munroNames[rand.Intn(282)]

	resp := GetMunroResponse{
		Munro:  munro,
		Index:  index,
		Next:   next,
		Prev:   prev,
		Random: random,
	}

	c.JSON(200, resp)
}

func (con *Controller) ListMunroNames(c *gin.Context) {

	var munroNames []string
	for k := range con.Munros {
		munroNames = append(munroNames, k)
	}
	c.JSON(200, munroNames)
}

func (con *Controller) ListMunros(c *gin.Context) {
	query := NewListQuery(c)

	var munros []Munro

	paginatedQuery := New(con.MongoClient.Collection).Context(con.MongoClient.Ctx).Limit(query.Limit).Page(query.Page).Filter(query.Filter).Sort(query.SortParam, query.SortIndex)

	if query.Selector != nil {
		paginatedQuery = paginatedQuery.Select(query.Selector)
	}

	paginatedMunros, err := paginatedQuery.Decode(&munros).Find()

	if err == mongo.ErrNoDocuments {
		// Do something when no record was found
		fmt.Println("record does not exist")
	} else if err != nil {
		log.Fatal(err)
	}

	c.JSON(200, paginatedMunros)
}
