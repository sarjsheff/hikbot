package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
)

var sdkFlag = flag.String("s", "", "Hikvision SDK")
var imageFlag = flag.String("i", "", "Docker registry")
var regFlag = flag.String("r", "", "Docker image")
var userFlag = flag.String("u", "", "Registry username")
var passwordFlag = flag.String("p", "", "Registry password")

func main() {
	flag.Parse()
	if *sdkFlag == "" {
		fmt.Println("Set hikvision sdk path (-s).")
		return
	}
	if *userFlag == "" || *passwordFlag == "" || *regFlag == "" || *imageFlag == "" {
		fmt.Println("Set docker registry (-r -i -u -p).")
		return
	}

	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	golang := client.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).From("golang:1.19").WithWorkdir("/src").
		WithDirectory("/src", client.Host().Directory(".")).
		WithDirectory("/hiksdk", client.Host().Directory(*sdkFlag)).
		WithEnvVariable("CGO_CXXFLAGS", "-I/hiksdk/incEn/").
		WithEnvVariable("CGO_LDFLAGS", "-L/hiksdk/lib -lhcnetsdk").
		WithEnvVariable("GO_OS", "linux").
		WithEnvVariable("GO_ARCH", "amd64").
		WithExec([]string{"go", "build"})

	golang.File("hikbot").Export(ctx, "./hikbot")

	secret := client.SetSecret("password", *passwordFlag)

	_, err = client.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("debian").
		WithDirectory("/hiksdk", client.Host().Directory(*sdkFlag)).
		WithFile("/bin/hikbot", golang.File("hikbot")).
		WithEnvVariable("LD_LIBRARY_PATH", "/hiksdk/lib").
		WithEntrypoint([]string{"/bin/hikbot"}).
		WithRegistryAuth(*regFlag, *userFlag, secret).
		Publish(ctx, *imageFlag)

	if err != nil {
		log.Println(err)
	}
}
