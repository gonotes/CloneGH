package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need organization name")
		return
	}

	orgName := os.Args[1]
	os.Mkdir(orgName, os.ModeDir)

	pageNum := 0
	for {
		var url string
		if pageNum > 0 {
			url = fmt.Sprintf("https://github.com/%s?page=%d", orgName, pageNum)
		} else {
			url = fmt.Sprintf("https://github.com/%s", orgName)
		}
		fmt.Println(url)
		if !getOnePage(orgName, url) {
			fmt.Println("No more pages")
			break
		}

		pageNum++
	}

}

func getOnePage(orgName, url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}

	defer resp.Body.Close()

	//match something like this:  <a href="/docker/libnetwork"
	mstr := fmt.Sprintf(" <a class=\"d-inline-block\" href=\"/%s/[[:alnum:]-_]+\"", orgName)

	re := regexp.MustCompile(mstr)
	body, err := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(body), "This organization has no more repositories.") {
		return false
	}
	//fmt.Println(string(body))

	var wg sync.WaitGroup
	repos := re.FindAllString(string(body), -1)
	for _, repo := range repos {
		left := strings.LastIndex(repo, "/")
		if left == -1 {
			continue
		}
		repoName := repo[left+1 : len(repo)-1]
		gitUrl := fmt.Sprintf("https://github.com/%s/%s.git", orgName, repoName)
		fmt.Println("Try to get code from:", gitUrl)
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			localDir := fmt.Sprintf("%s/%s", orgName, repoName)
			cmd := exec.Command("git", "clone", gitUrl, localDir)
			o, err := cmd.Output()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(o))
			wg.Done()
		}(&wg)
	}

	fmt.Println("Wait git clone exit...")
	wg.Wait()

	return true
}
