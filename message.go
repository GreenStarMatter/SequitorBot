package main

import (
	"encoding/json"
	"bytes"
	"fmt"
	"net/http"
	"bufio"
	"os"
	"strings"
	"errors"
	"math/rand"
	"time"
	"io"
	"regexp"
)
var (
	token     = readConf("config.conf", "TOKEN")
	channelID = readConf("config.conf", "CHANNEL_ID")
	apiURL    = "https://discord.com/api/v10/channels/" + channelID + "/messages"
)


type Author struct {
	Username string `json:"username"`
	ID       string `json:"id"`
	Bot      bool   `json:"bot"`
}

type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Author  Author `json:"author"`
}



func check(e error) {
    if e != nil {
        panic(e)
    }
}

func readConf(fileName string, param string) string {
	f, err := os.Open(fileName)
	check(err)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		configParam := strings.Split(line, " = ")[0]
		if configParam == "["+param+"]" {
			return strings.Split(line, " = ")[1]
		}
	}
	return ""
}


func postMessage(message string) {
	
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(message)))
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Message sent! Status:", resp.Status)
}

func readMessages() []Message {
	
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return []Message{}
	}
	defer resp.Body.Close()
	var messages []Message
	err = json.NewDecoder(resp.Body).Decode(&messages)
	if err != nil {
		fmt.Println("Decode error:", err)
		return []Message{}
	}
	return messages
}
func searchMessages(messageID string, messages []Message) (Message, error) {
	for _, message := range messages {
		if message.ID  == messageID {
			return message, nil
		}
	}
	return Message{}, errors.New("message not found: " + messageID)
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func connectMessagesAndCreateResponse(m1 Message, m2 Message) string {
	maxPrintLen := 10
	return fmt.Sprintf("Of course I know why you said '%s' %s.  It's because you were inextricably entangled in a quantum field with %s the moment they uttered '%s', or was it the other way around?",m1.Content[:maxInt(maxPrintLen, len(m1.Content))], m1.Author.Username, m2.Author.Username, maxInt(maxPrintLen, len(m2.Content)))
}

func possibleMessageConnections(m1 Message, m2 Message) string {
	//maxPrintLen := 10
	m1Summary := sendToOllama(m1.Content)
	m2Summary := sendToOllama(m2.Content)
	choices := []string {
			fmt.Sprintf("Of course I know why you said '%s' %s.  It's because you were inextricably entangled in a quantum field with %s the moment they uttered '%s', or was it the other way around?", m1Summary, m1.Author.Username, m2.Author.Username, m2Summary),
			fmt.Sprintf("Oh you don't know? Well it's because %s said '%s'.  As we all know this started an abnormal drying of the Permian Basin which caused the residents to say 'Wow, it certainly is dry today.' However it had the distinct effect of making %s come to the inevitable conlcusion '%s', or was it the other way around?", m1.Author.Username, m1Summary, m2.Author.Username, m2Summary),
			fmt.Sprintf("Sometimes it's simple. %s said '%s'.  %s immediately had a mid-mid-life crisis that they handled quite well until they relapsed at a crucial moment of a poorly made peanut butter and jelly sandwich with a vain uttering of '%s', or was it the other way around?", m1.Author.Username, m1Summary, m2.Author.Username, m2Summary),
			fmt.Sprintf("It's not easy to say this. There is rumors that '%s' was said by %s as a direct instigation to %s.  It was completely understandable that reliation would occur, but no one expected '%s', or was it the other way around?", m1Summary, m1.Author.Username, m2.Author.Username, m2Summary),
		}


	return choices[rand.Intn(len(choices))]
}

type ollamaRequest struct {
	Model string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func sendToOllama(m string) string {
	unhingedLevel := rand.Intn(9) +1
	reqBody := ollamaRequest{
		Model:  "llama2",
		Prompt: fmt.Sprintf("Explain what was said in a bad and funny way with an unhinged level of %d out of 10, 15 word sentence max, text only please and do not use double quotes: ", unhingedLevel) + m,
	}
	body, _ := json.Marshal(reqBody)

	// Send the request to Ollama
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var output string
	decoder := json.NewDecoder(resp.Body)
	for {
		var res ollamaResponse
		if err := decoder.Decode(&res); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		re := regexp.MustCompile(`[^\x20-\x7E]+`)
		output += strings.ReplaceAll(re.ReplaceAllString(res.Response, ""), "\"", "")
		if res.Done {
			break
		}
	}
	return output
}

func main() {
	rand.Seed(time.Now().UnixNano())
	messagesRead := readMessages()
	m1, err1 := searchMessages(readConf("config.conf", "MESSAGE_ID1"), messagesRead)
	if err1 != nil {
		panic(err1)
	}
	m2, err2 := searchMessages(readConf("config.conf", "MESSAGE_ID2"), messagesRead)
	if err2 != nil {
		panic(err2)
	}
	messageContent := possibleMessageConnections(m1, m2)
	message := fmt.Sprintf(`{"content": "%s"}`, messageContent)
	//fmt.Println(messageContent)
	//postMessage(messageContent)
	fmt.Println(message)
	postMessage(message)
}