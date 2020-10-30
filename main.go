package main

import (
	"net/http"
	"fmt"
	"regexp"
	"strings"
	"github.com/AlecAivazis/survey/v2"
	"github.com/PuerkitoBio/goquery"
	mapset "github.com/deckarep/golang-set"
)

var blockedTags = []interface{}{
	"#bhfyp",
	"#ol",
	"#follow",
	"#me",
	"#love",
	"#like",
}
var blockedTagsSet = mapset.NewSetFromSlice(blockedTags)

var requiredTags = []interface{} {
	"#tiktok",
	"#9gag",
	"#meme",
	"#memes",
	"#dailymemes",
	"#memesdaily",
	"#memepage",
}
var requiredTagsSet = mapset.NewSetFromSlice(requiredTags)

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

	foundTags := mapset.NewSet()

	// Gathering all tags from HTMLs
	for _, tag := range tags {
		go getTagsAtUrl(createUrlForHashtag(tag), dataChannel)
	}

	for  i := 0; i < len(tags); i++{
		tagList :=  <- dataChannel
		for _, t := range tagList{
			foundTags.Add(t)
		}
	}

	// Removing banned tags and constant tags
	foundTags = foundTags.Difference(blockedTagsSet)
	foundTags = foundTags.Difference(requiredTagsSet)
	
	tagsToReturn := make([]string, foundTags.Cardinality())
	for i, v := range foundTags.ToSlice() {
		tagsToReturn[i] = v.(string)
	}

	return tagsToReturn
}

func setToStrings(set mapset.Set) []string {
	tagsToReturn := make([]string, set.Cardinality())
	for i, v := range set.ToSlice() {
		tagsToReturn[i] = v.(string)
	}
	return tagsToReturn

}

func main() {

	var hashtagsToSearch = []string{}
	for true{
		tag := ""
		prompt := &survey.Input{
			Message: "Enter new hashtag to search (enter none to stop)",
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


	fmt.Printf("Hashtags: %s %s",  strings.Join(setToStrings(requiredTagsSet), " "), strings.Join(hashtagsToKeep, " "))

}
