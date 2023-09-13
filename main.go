package main

import (
    "bufio"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/chromedp/cdproto/dom"
    "github.com/chromedp/chromedp"
    "log"
    "os"
    "strings"
    "sync"
    "time"
)

// Matcher represents a key-value pair for matching content.
type Matcher struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    concurrency := flag.Int("p", 5, "number of concurrent executions")
    configFile := flag.String("config", "~/dominspect.json", "path to the JSON configuration file")
    flag.Parse()

    // Load matchers from the JSON file
    matchers, err := loadMatchers(*configFile)
    if err != nil {
        log.Fatalf("Error loading matchers from the configuration file: %v\n", err)
    }

    // Create a scanner to read URLs from stdin
    scanner := bufio.NewScanner(os.Stdin)
    var wg sync.WaitGroup

    // Create a channel to limit concurrency
    semaphore := make(chan struct{}, *concurrency)

    for scanner.Scan() {
        url := scanner.Text()
        if url == "" {
            continue // Skip empty lines
        }

        // Acquire a semaphore to control concurrency
        semaphore <- struct{}{}

        // Start a new goroutine for each URL
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            defer func() { <-semaphore }()

            // Initialize a controllable Chrome instance for each URL
            ctx, cancel := chromedp.NewContext(
                context.Background(),
            )

            // To release the browser resources when it is no longer needed
            defer cancel()

            var html string
            err := chromedp.Run(ctx,
                // Visit the target page
                chromedp.Navigate(url),
                // Wait for the page to load
                chromedp.Sleep(2000*time.Millisecond),
                // Extract the raw HTML from the page
                chromedp.ActionFunc(func(ctx context.Context) error {
                    // Select the root node on the page
                    rootNode, err := dom.GetDocument().Do(ctx)
                    if err != nil {
                        return err
                    }
                    html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
                    return err
                }),
            )
            if err != nil {
                log.Printf("Error while performing the automation logic for URL %s: %v\n", url, err)
                return
            }

            // Find and print key-value pairs in the HTML
            for _, matcher := range matchers {
                if strings.Contains(html, matcher.Value) {
                    fmt.Printf("%s - %s - %s\n", matcher.Key, matcher.Value, url)
                }
            }
        }(url)
    }

    if err := scanner.Err(); err != nil {
        log.Fatal("Error reading input:", err)
    }

    // Wait for all goroutines to finish
    wg.Wait()
    close(semaphore)
}

// loadMatchers loads matchers from a JSON file.
func loadMatchers(filePath string) ([]Matcher, error) {
    // Expand the user's home directory if ~ is used
    if strings.HasPrefix(filePath, "~/") {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return nil, err
        }
        filePath = strings.Replace(filePath, "~/", homeDir+"/", 1)
    }

    // Open and read the JSON file
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var matchers []Matcher
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&matchers); err != nil {
        return nil, err
    }

    return matchers, nil
}
