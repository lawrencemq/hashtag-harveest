package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PuerkitoBio/goquery"
)

var blockedTags = []string{
	"bhfyp",
	"ol",
	"follow",
	"me",
	"love",
	"like",
}

var requiredTags = []string{
	"tiktok",
	"9gag",
	"meme",
	"memes",
	"dailymemes",
	"memesdaily",
	"memepage",
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

	foundTags := map[string]bool{}

	// Adding requested tags
	for _, tag := range tags {
		foundTags[tag] = true
	}

	// Gathering all tags from HTMLs
	for _, tag := range tags {
		go getTagsAtUrl(createUrlForHashtag(tag), dataChannel)
	}

	for i := 0; i < len(tags); i++ {
		tagList := <-dataChannel
		for _, t := range tagList {
			foundTags[t[1:]] = true
		}
	}

	// Removing banned tags and constant tags
	for _, tag := range blockedTags {
		delete(foundTags, tag)
	}

	// Removing required tags as they're added in later
	for _, tag := range requiredTags {
		delete(foundTags, tag)
	}

	// Returning just the tags
	finalTags := make([]string, 0, len(foundTags))
	for tag := range foundTags {
		finalTags = append(finalTags, tag)
	}

	return finalTags
}

func main() {

	var hashtagsToSearch = []string{}
	for true {
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
		Message: "Which hashtags shall be kept:",
		Options: hashtagList,
	}
	survey.AskOne(prompt, &hashtagsToKeep)

	fmt.Printf("Hashtags: #%s #%s #%s", strings.Join(hashtagsToSearch, " #"), strings.Join(requiredTags, " #"), strings.Join(hashtagsToKeep, " #"))

}
