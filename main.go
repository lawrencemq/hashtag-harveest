package main

import (
	"net/http"
	"fmt"
	"regexp"
	"strings"
	"github.com/AlecAivazis/survey/v2"
	"github.com/PuerkitoBio/goquery"
)

var blockedTags  = map[string]bool{
	"#bhfyp": true,
	"#ol": true,
	"#follow": true,
	"#me": true,
	"#love": true,
	"#like": true,
}

func getTagsAtUrl(url string, tagChannel chan []string) {
	
	fmt.Printf("Getting %s...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	document, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
       panic(err)
    }

	found := document.Find(".tag-box")

	tagChannel <- strings.Fields(found.Text())
}

func createUrlForHashtag(tag string) string {
	template := "http://best-hashtags.com/hashtag/%s/"
	
	 // only want letters and numbers
	 reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	 if err != nil {
		 panic(err)
	 }
	 cleanedTag := reg.ReplaceAllString(tag, "")
	
	return fmt.Sprintf(template, cleanedTag)
}


func getDataForHashtags(tags []string) []string {

	dataChannel := make(chan []string)

	foundTags := make(map[string]bool)

	// Gathering all tags from HTMLs
	for _, tag := range tags {
		go getTagsAtUrl(createUrlForHashtag(tag), dataChannel)
	}

	for  i := 1; i < len(tags); i++{
		tagList :=  <- dataChannel
		for _, t := range tagList{
			foundTags[t] = true
		}
	}

	// Removing banned tags
	for t := range blockedTags{
		if _, ok := foundTags[t]; ok {
			delete(foundTags, t)
		}
	}

	finalTags := make([]string, 0, len(foundTags))
    for k := range foundTags {
        finalTags = append(finalTags, k)
    }

	return finalTags
}

func main() {

	var hashtagsToSearch = []string{}
	for true{
		tag := ""
		prompt := &survey.Input{
			Message: "ping",
		}
		survey.AskOne(prompt, &tag)
		if len(tag) == 0 {
			break
		}
		hashtagsToSearch = append(hashtagsToSearch, tag)
	}
	
	
	hashtagList := getDataForHashtags(hashtagsToSearch)

	hashtagsToKeep := []string{}
	prompt := &survey.MultiSelect{
		Message: "Which hashtags shal be kept:",
		Options: hashtagList,
	}
	survey.AskOne(prompt, &hashtagsToKeep)

}
