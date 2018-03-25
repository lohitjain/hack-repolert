package main

import (
    "encoding/json"
    "github.com/gorilla/mux"
    "log"
    "net/http"
    "strconv"
    "io/ioutil"
    "fmt"
    "bytes"
    "strings"
)

// The person Type (more like an object)
type Event struct {
    ID          string   `json:"id"`
    Keywords    []string   `json:"keywords"`
    NumPeople   int `json:"numPeople"`
    EventType   string   `json:"type"`
    EventStatus int `json:"status"`
    Location     *Address `json:"address"`
    Severity    int `json:"severity"`
}

type InputEvent struct {
    Location *Address `json:"address"`
    Comment string `json:"comments"`
    //EventType int `json:"type"` 
}

type StatusChange struct {
    EventStatus int `json:"status"`
}

type EventUpdate struct {
    Comment string `json:"comments"`
}

type WatsonOutput struct {
    Keywords []string
    Category string
}

type WatsonAPIOutput struct {
   Usage struct {
       TextUnits      int `json:"text_units"`
       TextCharacters int `json:"text_characters"`
       Features       int `json:"features"`
   } `json:"usage"`
   Language string `json:"language"`
   Keywords []struct {
       Text      string  `json:"text"`
       Relevance float64 `json:"relevance"`
   } `json:"keywords"`
   Categories []struct {
       Score float64 `json:"score"`
       Label string  `json:"label"`
   } `json:"categories"`
}

type WatsonAPIInput struct {
   Text     string `json:"text"`
   Features struct {
       Keywords struct {
           Limit int `json:"limit"`
       } `json:"keywords"`
       Categories struct {
       } `json:"categories"`
   } `json:"features"`
}

const (
        Fire        = 0
        EMS         = 1
        Police      = 2
    )

const (
        NO_RESPONSE = 0
        RESPONDING  = 1
        RESPONDED   = 2
    )

type Address struct {
    Latitude  string `json:"latitude"`
    Longitude string `json:"longitude"`
}

var events []Event
var gid int


//---------------------------------------------------
// Application APIs
//---------------------------------------------------

func lookupEvent(id string) Event {
    for _, item := range events {
        if item.ID == id {
            return item
        }
    }
    return Event{}
}


//----------------------------------------------------
// RESTful APIs
//----------------------------------------------------

func GetEvents(w http.ResponseWriter, r *http.Request) {
   // params := mux.Vars(r)
    json.NewEncoder(w).Encode(events)
    //get all events
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
    //params := mux.Vars(r)
    var event InputEvent
    _ = json.NewDecoder(r.Body).Decode(&event)

    var watsonInputString string 
    watsonInputString = event.Comment
    //var eType int := event["type"]
    var location *Address = event.Location

    var id string = strconv.Itoa(gid)
    gid = gid+1
    var watsonOutput WatsonOutput = watsonUnderstandComment(watsonInputString)

    var keywords []string = watsonOutput.Keywords
    var eType string = watsonOutput.Category

    var storeEvent Event = Event{id, keywords, 1, eType, 0, location, 1}

    events = append(events, storeEvent)
    json.NewEncoder(w).Encode(events)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)

    var eventUpdate EventUpdate
    _ = json.NewDecoder(r.Body).Decode(&eventUpdate)

    var event Event

    for _, item := range events {
        if item.ID == params["id"] {
            event = item
        }
    }

    var watsonInputString string 
    watsonInputString = eventUpdate.Comment

    var watsonOutput WatsonOutput = watsonUnderstandComment(watsonInputString)

    var keywords []string = watsonOutput.Keywords
    //var eType string = watsonOutput.Category

    for _, item := range keywords {
        event.Keywords = append(event.Keywords, item)
    }

    json.NewEncoder(w).Encode(events)

}

func GetEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    json.NewEncoder(w).Encode(lookupEvent(params["id"]))
}

func GetResponderEvents(w http.ResponseWriter, r *http.Request) {
    //params := mux.Vars(r)
    json.NewEncoder(w).Encode(events)
}

func GetResponderEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    json.NewEncoder(w).Encode(lookupEvent(params["id"]))
}

func UpdateResponderEventStatus(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var updateEvent StatusChange
    _ = json.NewDecoder(r.Body).Decode(&updateEvent)
    for i, item := range events {
        if item.ID == params["id"] {
            item.EventStatus = updateEvent.EventStatus
            events[i]= item
            json.NewEncoder(w).Encode(item)
            return
        }
    }
    json.NewEncoder(w).Encode(&Event{})
}

func DeleteResponderEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    for index, item := range events {
        if item.ID == params["id"] {
            events = append(events[:index], events[index+1:]...)
            break
        }
        json.NewEncoder(w).Encode(events)
    }
}




// main function to boot up everything
func main() {
    router := mux.NewRouter()
    router.HandleFunc("/public/events", GetEvents).Methods("GET")
    router.HandleFunc("/public/events", CreateEvent).Methods("POST")
    router.HandleFunc("/public/events/{id}", UpdateEvent).Methods("PUT")
    router.HandleFunc("/public/events/{id}", GetEvent).Methods("GET")
    router.HandleFunc("/responder/events", GetResponderEvents).Methods("GET")
    router.HandleFunc("/responder/events/{id}", GetResponderEvent).Methods("GET")
    router.HandleFunc("/responder/events/{id}", UpdateResponderEventStatus).Methods("PUT")
    router.HandleFunc("/responder/events/{id}", DeleteResponderEvent).Methods("DELETE")
    gid = 100
    log.Fatal(http.ListenAndServe(":8000", router))
}

/*func watsonUnderstandComment(watsonInputString string) WatsonOutput {
    var keys []string
    keys = append(keys, "asdf")

    cat := "qwe"
    return WatsonOutput{keys, cat}
}
*/

func createWatsonInput(comments string) string {

   inputSample := `{"text": "SET_THIS_TEXT", "features": {"keywords": {"limit": 5 }, "categories": {} } }`
   outputString := strings.Replace(inputSample,"SET_THIS_TEXT", comments,-1)
   return outputString
}

func watsonUnderstandComment(comments string) WatsonOutput {

   var watsonInputString = []byte(createWatsonInput(comments))
   client := &http.Client{}
   watsonUrl := "https://gateway.watsonplatform.net/natural-language-understanding/api/v1/analyze?version=2018-03-16"
   req, err := http.NewRequest("POST", watsonUrl, bytes.NewBuffer(watsonInputString))
   if req==nil {
           fmt.Println("hello world2-err %s\n", err);
       }
   req.SetBasicAuth("50954fd3-ce24-4b57-b303-c5fb4737d788", "5hwLXHjD8ciG")
   req.Header.Set("Content-Type", "application/json")
   resp, err := client.Do(req)
   if err != nil {
       fmt.Printf("The HTTP request failed with error %s\n", err)
   } 
   data, _ := ioutil.ReadAll(resp.Body)
   fmt.Println(string(data))
   
   watsonOutputJSONString := string(data)
   byt:= []byte(watsonOutputJSONString)
   //var dat map[string]interface{}
   res := WatsonAPIOutput{}
   if err := json.Unmarshal(byt, &res); err != nil {
       panic(err)
   }
   //fmt.Println(dat)

   var keywords []string
   keywords = append(keywords,res.Keywords[0].Text)
   keywords = append(keywords,res.Keywords[1].Text)

   categories := res.Categories[0].Label

   return WatsonOutput{keywords,categories}
   
}
/*func createWatsonInpute(comments string){

   var Categories string
   var Keywords []string
   
}

func watsonUnderstandComment(watsonInputString string) WatsonOutput {
   client := &http.Client{}
   watsonUrl := "https://gateway.watsonplatform.net/natural-language-understanding/api/v1/analyze?version=2018-03-16"
   req, err := http.NewRequest("POST", watsonUrl, bytes.NewBuffer(string))
   if req==nil {
           fmt.Println("hello world2-err %s\n", err);
       }
   req.SetBasicAuth("50954fd3-ce24-4b57-b303-c5fb4737d788", "5hwLXHjD8ciG")
   req.Header.Set("Content-Type", "application/json")
   resp, err := client.Do(req)
   if err != nil {
       fmt.Printf("The HTTP request failed with error %s\n", err)
   } else {
       data, _ := ioutil.ReadAll(resp.Body)
       fmt.Println(string(data))
   }
   watsonOutputJSONString = string(data)
   byt:= []byte(watsonOutputJSONString)
   //var dat map[string]interface{}
   res := WatsonAPIOutput{}
   if err := json.Unmarshal(byt, &res); err != nil {
       panic(err)
   }
   //fmt.Println(dat)

   var keywords []string
   keywords = append(keywords,res.Keywords[0].Text]
   keywords = append(keywords,res.Keywords[1].Text]

   categories= res.Categories[0].Label

   return watsonOutput(keywords,categories)
   
}*/
