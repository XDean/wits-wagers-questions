package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/xdean/goex/xecho"
	"github.com/xdean/goex/xgo"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Question struct {
	Q, A string
}

var (
	contentPath = flag.String("content", "", "path to question files")
	port        = flag.Int("port", 11076, "Command Port")
)

func main() {

	flag.Parse()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(xecho.BreakErrorRecover())
	e.Use(middleware.CORS())

	e.GET("/qs", func(c echo.Context) error {
		return c.JSON(200, allQuestionSuites())
	})

	e.GET("/qs/:name/random", func(c echo.Context) error {
		name := c.Param("name")
		qs, err := getQuestionSuite(name)
		xecho.MustNoError(err)
		return c.JSON(200, qs[rand.Intn(len(qs))])
	})

	e.GET("/qs/:name/:index", func(c echo.Context) error {
		name := c.Param("name")
		index := IntParam(c, "index")
		qs, err := getQuestionSuite(name)
		xecho.MustNoError(err)
		if len(qs) <= index {
			return c.JSON(400, xecho.M(fmt.Sprintf("Index out of bound: len %d got %d", len(qs), index)))
		}
		return c.JSON(200, qs[index])
	})

	log.Fatal(e.Start(":" + strconv.Itoa(*port)))
}

func allQuestionSuites() []string {
	infos, err := ioutil.ReadDir(*contentPath)
	xgo.MustNoError(err)
	res := make([]string, 0)
	for _, v := range infos {
		if filepath.Ext(v.Name()) == ".json" {
			res = append(res, v.Name()[:len(v.Name())-5])
		}
	}
	return res
}

func getQuestionSuite(name string) ([]Question, error) {
	f := filepath.Join(*contentPath, name+".json")
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return nil, echo.NewHTTPError(400, "No Such Question Suite: "+name)
	}
	s, err := os.Open(f)
	xgo.MustNoError(err)
	defer s.Close()
	qs := make([]Question, 0)
	err = json.NewDecoder(s).Decode(&qs)
	xecho.MustNoError(err)
	return qs, nil
}

func IntParam(c echo.Context, name string) int {
	param := c.Param(name)
	if i, err := strconv.Atoi(param); err == nil {
		return i
	} else {
		xecho.MustNoError(echo.NewHTTPError(http.StatusBadRequest, "Unrecognized param '"+name+"': "+param))
		return 0
	}
}
