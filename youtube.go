//TODO
// 1- Build model for basic youtube API answer
// We need to put the data into model using JSON Unmarshal

// 2- Test simple channel request based on user handle (we only need channel ID)

// 3- Now we need to get the content inside Items in each Answer

// 4- Get Playlists list from a channel

// 5- Get Videos(PlayListItems) lists from playlist

// DONE
//
// 6- Ignore empty results

// 7- Handle Pagination (not today :))
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// Global variables
var (
	YOUTUBE_API_KEY = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	YOUTUBE_API_ENDPOINT = "https://www.googleapis.com/youtube/v3/"

	TEST_CHANNEL_HANDLE = "TotalHalibut"
)

/////////////////////////////
//////////// MODELS
///////////////////////////

// model for a generic youtube answer
// We only define the fields we want to capture
// from the json answer from the API

type PageInfo struct {
	TotalResults   int
	ResultsPerPage int
}

type YoutubeAnswer struct {
	Kind  string
	Etag  string
	Items []Item
	PageInfo

	// Will be used to get the actual we need (resources)
	// Now we need to get the content inside Items in each Answer
	// Let's build a type(model) for one Item
	// interface{} means any type like void* in C
	// Now we can use the ChannelItem type
}

type PlayList struct {
}

type ChannelItem struct {
}

type ResourceId struct {
	VideoId string
}

type PlayListItemSnippet struct {
	ResourceId ResourceId
}

type PlayListItem struct {
	Snippet PlayListItemSnippet
}

// Our container type Item
// We will embed other type of Items here
// to extract data we want from our resources with json.Unmarshal
// Unmarshal should set to nil json fields which are not defined
// There might be a cleaner way to do this
// Ref on embedding: https://github.com/luciotato/golang-notes/blob/master/OOP.md#golang-embedding-is-akin-to-multiple-inheritance-with-non-virtual-methods

type Item struct {
	Id string
	ChannelItem
	PlayList
	PlayListItem
}

////////////////////////////
////// API Functions
///////////////////////////

// Helper function to build our youtube query URL
func buildUrl(resource string, params map[string]string) string {

	// we build a url.Values object from a map of key,values

	queryParams := url.Values{}

	// Automatically add the API key params to all our calls
	queryParams.Set("key", YOUTUBE_API_KEY)

	for k, v := range params {
		queryParams.Set(k, v)
	}

	// Build URL
	queryUrl := fmt.Sprint(YOUTUBE_API_ENDPOINT, resource, "?", queryParams.Encode())
	return queryUrl
}

// We can make this function more generic and reuse it for other resources
// Since youtube API sends the actual data inside ITEMS we make a func
// that returns the list of Items based on the resource and params we pass
// This func returns a list of any kind of Item type (channel, playlist, video ...)

func getResource(resource string, params map[string]string) []Item {

	queryUrl := buildUrl(resource, params)
	resp, err := http.Get(queryUrl)

	if err != nil {
		log.Fatal(err)
	}

	data, _ := ioutil.ReadAll(resp.Body)

	answer := &YoutubeAnswer{}
	jsonErr := json.Unmarshal(data, answer)

	if jsonErr != nil {
		log.Fatal(err)
	}

	// We need to make an type embedding so we can have polymorphism
	// Interfaces are used to implement methods , my mistake
	// We want to return a list of any "kind" if Item

	return answer.Items
}

func getChannelId(channelHandle string) string {

	// Use our buildUrl func
	params := map[string]string{
		"key":         YOUTUBE_API_KEY,
		"forUsername": channelHandle,
		"part":        "id",
	}

	queryUrl := buildUrl("channels", params)

	// Test request
	resp, err := http.Get(queryUrl)
	if err != nil {
		log.Fatal(err)
	}

	data, _ := ioutil.ReadAll(resp.Body)
	answer := &YoutubeAnswer{}
	jsonErr := json.Unmarshal(data, answer)

	if jsonErr != nil {
		log.Fatal(err)
	}

	return answer.Items[0].Id
}

///////////////////////////////////

func main() {

	// Get channel data
	channelParams := map[string]string{
		"forUsername": TEST_CHANNEL_HANDLE,
		"part":        "id",
	}
	channel := getResource("channels", channelParams)[0]

	// Use channel Id to get list of playlists

	playlistsParams := map[string]string{
		"part":      "id",
		"channelId": channel.Id,
	}

	playlists := getResource("playlists", playlistsParams)

	// Step 2 get list of playlists from channel

	fmt.Printf("Id for channel %s is: %s\n", TEST_CHANNEL_HANDLE, channel.Id)
	fmt.Println("_________Playlists IDS___________\n")

	// Step 3 get list of videos inside each playlist
	// Videos are PlayListItem objects
	// We want the id of videos also (part=id,snippet)

	for _, pl := range playlists {
		fmt.Printf("\n%+v", pl)

		fmt.Printf("playlist: %s\n", pl.Id)

		playListItemsParams := map[string]string{
			"part":       "id,snippet",
			"playlistId": pl.Id,
		}

		playListItems := getResource("playlistItems", playListItemsParams)

		fmt.Printf("video IDs for playlist %s:\n", pl.Id)

		for _, plI := range playListItems {
			fmt.Printf("video id : %s\n", plI.PlayListItem.Snippet.ResourceId.VideoId)

		}

	}

	return
}
