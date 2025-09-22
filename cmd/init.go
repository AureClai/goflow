package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goflow",
	Short: "GoFlow is a front-end web framework for Go and WebAssembly.",
	Long: `GoFlow provides a CLI to initialize, build, and serve Go-based
front-end applications that compile to WebAssembly.`,
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initializes a new GoFlow project.",
	Long: `Creates a new directory with the specified project name and populates it with
the basic structure and files needed to get started with a GoFlow application.`,
	Args: cobra.ExactArgs(1), // Ensures exactly one argument (the project name) is passed
	Run:  runInit,
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the GoFlow application into a Wasm module.",
	Long: `Compiles the Go source code into a WebAssembly module (app.wasm) and
copies the necessary wasm_exec.js file. This command should be run from
the root of a GoFlow project.`,
	Run: runBuild,
}

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Runs a local development server for the GoFlow application.",
	Long: `Starts a static file server on port 8080 to serve the application.
It is recommended to run 'goflow build' before starting the dev server.`,
	Run: runDev,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(devCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// runInit is the function executed when the 'init' command is called.
func runInit(cmd *cobra.Command, args []string) {
	projectName := args[0]

	fmt.Printf("🚀 Initializing new GoFlow project: %s\n", projectName)

	// Create project directory
	if err := os.Mkdir(projectName, 0755); err != nil {
		fmt.Printf("❌ Error creating project directory: %v\n", err)
		os.Exit(1)
	}

	// Define files to create with their content
	filesToCreate := map[string]string{
		"main.go":    mainGoTemplate,
		"index.html": indexHTMLTemplate,
		"go.mod":     goModTemplate(projectName),
		"README.md":  readmeTemplate(projectName),
		".gitignore": gitignoreTemplate,
	}

	for fileName, content := range filesToCreate {
		filePath := filepath.Join(projectName, fileName)
		err := os.WriteFile(filePath, []byte(strings.TrimSpace(content)), 0644)
		if err != nil {
			fmt.Printf("❌ Error creating file %s: %v\n", fileName, err)
			// Cleanup: attempt to remove created directory
			os.RemoveAll(projectName)
			os.Exit(1)
		}
		fmt.Printf("✅ Created %s\n", filePath)
	}

	fmt.Printf("\n🎉 Project '%s' created successfully!\n\n", projectName)
	fmt.Println("Next steps:")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Println("  2. Build the application: 'goflow build'")
	fmt.Println("  3. Start the dev server: 'goflow dev'")
	fmt.Println("  4. Open http://localhost:8080 in your browser.")
}

// runBuild handles the logic for the 'goflow build' command.
func runBuild(cmd *cobra.Command, args []string) {
	// Check if we are in a goflow project
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Println("❌ No main.go file found. Are you in a GoFlow project directory?")
		os.Exit(1)
	}

	fmt.Println("Building Go code to WebAssembly...")

	// Set environment variables for the build command.
	buildCmd := exec.Command("go", "build", "-o", "app.wasm", ".")
	buildCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	// Run the build command.
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("❌ Build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Build successful.")

	// Copy the wasm_exec.js file.
	if err := copyWasmExec(); err != nil {
		fmt.Printf("❌ Failed to copy wasm_exec.js: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Copied wasm_exec.js.")
	fmt.Println("\nBuild complete. You can now serve the directory using 'goflow dev'")
}

// runDev handles the logic for the 'goflow dev' command.
func runDev(cmd *cobra.Command, args []string) {
	// Check if the build artifacts exist
	if _, err := os.Stat("app.wasm"); os.IsNotExist(err) {
		fmt.Println("⚠️ app.wasm not found. Did you run 'goflow build' first?")
	}

	port := "8080"
	addr := ":" + port
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	fmt.Printf("Starting server on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// copyWasmExec finds and copies the wasm_exec.js file.
func copyWasmExec() error {
	goRoot := runtime.GOROOT()
	if goRoot == "" {
		return fmt.Errorf("GOROOT environment variable is not set")
	}

	srcPath := filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	destPath := "wasm_exec.js"

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("could not open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("could not create destination file %s: %w", destPath, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("could not copy file contents: %w", err)
	}
	return nil
}

// --- Templates ---

const mainGoTemplate = `
package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("Go Wasm app initialized.")

	// Get the document object
	document := js.Global().Get("document")
	if !document.Truthy() {
		fmt.Println("Could not get document object")
		return
	}

	// Get the app container
	appContainer := document.Call("getElementById", "app")
	if !appContainer.Truthy() {
		fmt.Println("Could not find element with id 'app'")
		return
	}

	// Create a new element
	h1 := document.Call("createElement", "h1")
	h1.Set("textContent", "Hello, GoFlow! 🚀")

	// Append the new element to the container
	appContainer.Call("appendChild", h1)

	// Keep the Go program running
	<-make(chan bool)
}
`

const indexHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoFlow App</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background-color: #f0f2f5; }
        #app { text-align: center; }
    </style>
</head>
<body>
    <div id="app">
        <h2>Loading WebAssembly...</h2>
		<p>If you see this message, the Go Wasm module is loading or has failed to load. Check the browser console for errors.</p>
    </div>

    <!-- The JS glue file provided by the Go installation -->
    <script src="wasm_exec.js"></script>
    <script>
        if (!WebAssembly.instantiateStreaming) { // polyfill
            WebAssembly.instantiateStreaming = async (resp, importObject) => {
                const source = await (await resp).arrayBuffer();
                return await WebAssembly.instantiate(source, importObject);
            };
        }

        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("app.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        }).catch((err) => {
            console.error("Wasm instantiation failed:", err);
			const appDiv = document.getElementById('app');
			appDiv.innerHTML = '<h2 style="color: red;">Error</h2><p>Failed to load WebAssembly module. Check console.</p>';
        });
    </script>
</body>
</html>
`

func goModTemplate(projectName string) string {
	return fmt.Sprintf(`
module %s

go 1.22
`, projectName)
}

func readmeTemplate(projectName string) string {
	return fmt.Sprintf(`
# %s

This project was generated by the GoFlow CLI.

## Development

`, projectName)
}

const gitignoreTemplate = `
# Compiled Wasm file
app.wasm

# JS glue file
wasm_exec.js
`
