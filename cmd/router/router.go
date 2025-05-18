package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

type RouteData struct {
	Route      string
	RouteLower string
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func main() {
	routePtr := flag.String("route", "", "Name of the route (e.g., ping, echo)")
	flag.Parse()

	if *routePtr == "" {
		log.Fatal("Route name is required. Use -route flag.")
	}

	routeLower := strings.ToLower(*routePtr)
	routeCap := Capitalize(routeLower)

	data := RouteData{
		Route:      routeCap,
		RouteLower: routeLower,
	}

	controllerStubPath := filepath.Join("stubs", "controller.stub")
	routerStubPath := filepath.Join("stubs", "router.stub")

	controllerStubBytes, err := os.ReadFile(controllerStubPath)
	if err != nil {
		log.Fatalf("Error reading controller stub: %v", err)
	}
	routerStubBytes, err := os.ReadFile(routerStubPath)
	if err != nil {
		log.Fatalf("Error reading router stub: %v", err)
	}

	controllerTmpl, err := template.New("controller").Parse(string(controllerStubBytes))
	if err != nil {
		log.Fatalf("Error parsing controller stub: %v", err)
	}
	routerTmpl, err := template.New("router").Parse(string(routerStubBytes))
	if err != nil {
		log.Fatalf("Error parsing router stub: %v", err)
	}

	controllerOutPath := filepath.Join("..", "app", "controller", routeLower+"_controller.go")
	routerOutPath := filepath.Join("..", "app", "router", "routes", routeLower+"_router.go")

	controllerFile, err := os.Create(controllerOutPath)
	if err != nil {
		log.Fatalf("Error creating controller output file: %v", err)
	}
	defer controllerFile.Close()

	routerFile, err := os.Create(routerOutPath)
	if err != nil {
		log.Fatalf("Error creating router output file: %v", err)
	}
	defer routerFile.Close()

	if err := controllerTmpl.Execute(controllerFile, data); err != nil {
		log.Fatalf("Error executing controller template: %v", err)
	}
	if err := routerTmpl.Execute(routerFile, data); err != nil {
		log.Fatalf("Error executing router template: %v", err)
	}

	fmt.Printf("Generated route files:\n - %s\n - %s\n", controllerOutPath, routerOutPath)
}
