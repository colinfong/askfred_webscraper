package main

/************
AskFred Webscraper
Goal: Web scrape public bay area fencing tournament results on AskFred.net to track the progress
of different fencers over several seasons.
************/

/*
TODO:
-Run Goroutines on each link
-Fix text node parsing
-Translate data to JSON
-Send over to RoR app analysis
-Grab previous season competition data
*/

import (
    "fmt"
    "net/http"
    "io"
    "io/ioutil"
    "log"
    "os"
    "golang.org/x/net/html"
    "golang.org/x/net/html/atom"
    "github.com/yhat/scrape"
    "strings"
)

//Check errors
func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

//Print HTML body
func printHTMLBody(body io.Reader) {
    bytes, _ := ioutil.ReadAll(body)
    fmt.Println(string(bytes))
}

//Print HTML body to file
func printHTMLBodyToFile(body io.Reader, fileDest string) {
    bytes, err := ioutil.ReadAll(body)
    checkErr(err)
    //ex: /home/colin/Desktop/go_webscraper/EGFTourney.html
    f, err := os.Create(fileDest)
    defer f.Close()
    checkErr(err)
    f.Write(bytes)
    fmt.Println(string(bytes))
}

//Get HTML body of URL
func getHTMLBody(url string) io.Reader {
    fmt.Printf("Retrieving HTML body of: %s\n", url)
    resp, err := http.Get(url)
    checkErr(err)
    return resp.Body
}

//All tournament tables have their attributs class="box" and width="500"
func gatherTournamentTables(node *html.Node) bool {
    if node.DataAtom == atom.Table && node.Parent != nil {
        return scrape.Attr(node, "class") == "box" && scrape.Attr(node, "width") == "500"
    }
    return false
}

//Gather table rows
func gatherTableRows(node *html.Node) bool {
    if node.DataAtom == atom.Tr {
        return true
    }
    return false
}

//Helper function to print nodes
func printNode(node *html.Node) {
    fmt.Printf("Node Type is: %s\n", node.Type)
    fmt.Printf("DataAtom is: %s\n", node.DataAtom)
    fmt.Printf("Data is: %s\n", node.Data)
    fmt.Printf("Namespace is: %s\n", node.Namespace)
    fmt.Println(node.Attr)
}

func rowsToData(nodes []*html.Node) ([]string, [][]string) {
    tInfo := make([]string, 0, 4)
    placings := make([][]string, 0, len(nodes)-2)
    for num := range placings {
        placings[num] = make([]string, 0, 5)
    }
    for i, node := range nodes {
        switch i {
        //First row contains: Event, # competitors, rating result, Pool & DE link
        case 0:
            //TODO: figure out why Tr->th->a scoping is so bad
            //printNode(node)
            textNodes := scrape.FindAll(node, func(n *html.Node) bool { return n.Type == html.TextNode })
            for _, tNode := range textNodes {
                fmt.Println("!",strings.TrimSpace(html.UnescapeString(tNode.Data)),"!")
            }
            //fmt.Println(scrape.Text(node))        
            //tInfo = append(tInfo, anchor.Data)
        //Second contains header for table - ignore
        case 1:
            //Skip
        //Placings contain: Place, Fencer, Club, Rating, Rating Earned
        default:
            for c := node.FirstChild; c != nil; c = c.NextSibling {
                printNode(c)
            }
            //TODO: text nodes print out oddly, figure out how to access data
            //Text nodes aren't treated the same as element nodes
            
        }
    }
    return tInfo, placings
}

func urlTournamentsToJSON(url string) string {
    body := getHTMLBody(url)
    //printHTMLBody(body)
    //printHTMLBodyToFile(body, "/home/colin/Desktop/go_webscraper/EGFTourney.html")
    //TODO: defer body.Close()?
    root, err := html.Parse(body)
    checkErr(err)
    //Find all tournaments on page
    tournamentTables:= scrape.FindAll(root, gatherTournamentTables)
    fmt.Printf("%d tournaments found \n", len(tournamentTables))
    for i, tournament := range tournamentTables {
        
        //Find all table rows
        tableRows := scrape.FindAllNested(tournament, gatherTableRows)
        fmt.Printf("There are %d rows in tournament %d\n", len(tableRows), i)
        tInfo, placings := rowsToData(tableRows) //tInfo, placings := rowsToData(tableRows)
        fmt.Println(tInfo)
        fmt.Println(placings)
    }
    return ""
}

func main() {
    urls := []string{
        "https://askfred.net/Results/results.php?tournament_id=37980",
        "https://askfred.net/Results/results.php?tournament_id=37977",
        "https://askfred.net/Results/results.php?tournament_id=37956",
        "https://askfred.net/Results/results.php?tournament_id=37912"
    }
    for url := range urls {
        go urlTournamentsToJSON(url)
    }
}