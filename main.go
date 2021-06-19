package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
	"io/ioutil"
	"github.com/valyala/fasthttp"
)

func main() {
	Clear()
	tokens := readTokens()
	setTitle(fmt.Sprintf("token checker | %v tokens", len(tokens)))
	fmt.Print(fmt.Sprintf("Start checking %v tokens? [Y/N] ", len(tokens)))
	var confirm string
	fmt.Scan(&confirm)
	if "y" == confirm || "Y" == confirm {
		StartCheck()
	} else {
		cmd := exec.Command("cmd", "/c", "exit")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

}

func Route(Method, Endpoint string) (req *fasthttp.Request, resp *fasthttp.Response) {
	req = fasthttp.AcquireRequest()
	req.Header.SetMethod(Method)
	req.SetRequestURI(fmt.Sprintf("%s%s", BASE_URL, Endpoint))
	resp = fasthttp.AcquireResponse()
	return req, resp
}

func Clear() {
	cmd := exec.Command("cmd", "/c", "cls")
    cmd.Stdout = os.Stdout
    cmd.Run()
}

func setTitle(title string) {
	cmd := exec.Command("cmd", "/C", "title", title)
	cmd.Stdout = os.Stdout
	cmd.Run()
}


func dumpTokens(tokens []string, filename string) {
	routines := make(chan struct{}, 35)
	var wg sync.WaitGroup
	for x := 0; x < len(tokens); x++ {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			routines <- struct{}{}
			err := ioutil.WriteFile("valid.txt", fmt.Sprintf("%s\n", token), 0644)
			if err != nil {
				return
			}
			<-routines
		}(tokens[x])
	}
	wg.Wait()
}

func StartCheck() {
	tokens := readTokens()
	var wg sync.WaitGroup
	goroutines := make(chan struct{}, 35)
	for i := 0; i < len(tokens); i++ {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			goroutines <- struct{}{}
			data := validateToken(token)
			if data == "invalid" {
				fmt.Printf("invalid token => %s\n", token)
			} else if data == "user" {
				fmt.Printf("user token => %s\n", token)
				valid = append(valid, token)
			} else if data == "bot" {
				fmt.Printf("bot token => %s\n", token)
				valid = append(valid, token)
			}
		}(tokens[i])
	}
	wg.Wait()
	fmt.Printf("finished checking %v tokens\n", len(tokens))
	if len(valid) != 0 {
		fmt.Printf("now dumping %v", len(valid))
	    dumpTokens(valid, "valid.txt")
	}
}

func readTokens() []string {
	file, err := os.Open("tokens.txt")
	if err != nil {
		fmt.Println(err)
	}
	scanned := bufio.NewScanner(file)
	scanned.Split(bufio.ScanLines)
	var tokens []string
	for scanned.Scan() {
		tokens = append(tokens, scanned.Text())
	}
	file.Close()
	return tokens
}

func validateToken(token string) string {
	req, resp := Route("GET", "/users/@me")
	req.Header.Set("Authorization", token)
	err := fasthttp.Do(req, resp)
	if err != nil {
		return "invalid"
	}
	if resp.StatusCode() == 204 || resp.StatusCode() == 201 ||  resp.StatusCode() == 200 {
		return "user"
	} else if resp.StatusCode() == 429 {
		time.Sleep(time.Second)
		return validateToken(token)
	} else {
		reqs, resps := Route("GET", "/users/@me")
		reqs.Header.Set("Authorization", fmt.Sprintf("Bot %s", token))
		errs := fasthttp.Do(reqs, resps)
		if errs != nil {
			return "invalid"
		}
		if resps.StatusCode() == 204 || resps.StatusCode() == 204 || resps.StatusCode() == 200 {
			return "bot"
		} else if resps.StatusCode() == 429 {
			time.Sleep(time.Second)
		    return validateToken(token)
		} else {
			return "invalid"
		}
	}
}





var (
	BASE_URL = "https://discord.com/api/v9"
	valid []string
)
