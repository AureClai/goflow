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
//go:build js && wasm

package main

import (
	"fmt"
	"github.com/AureClai/goflow/component"
	"github.com/AureClai/goflow/renderer"
)

func main() {
	fmt.Println("GoFlow app initialized! 🚀")

	// Create renderer
	r := renderer.NewRenderer("app")

	// Create the welcome page
	app := createWelcomePage()

	// Render the app
	r.Render(app.Render())

	// Keep the program running
	<-make(chan bool)
}

func createWelcomePage() *component.Container {
	// Main container
	page := component.NewContainer().SetClass("welcome-page")

	// Header section
	header := component.NewContainer().SetClass("header-section")
	
	// Logo and title
	logo := component.NewText("🔄").SetClass("logo")
	title := component.NewHeading("Welcome to GoFlow!", 1).SetClass("main-title")
	subtitle := component.NewParagraph("Your Go WebAssembly application is running successfully!").SetClass("subtitle")

	header.AddChild(logo)
	header.AddChild(title)
	header.AddChild(subtitle)

	// Features section
	features := component.NewContainer().SetClass("features-section")
	featuresTitle := component.NewHeading("What's Next?", 2).SetClass("section-title")
	features.AddChild(featuresTitle)

	// Feature list
	featureList := component.NewList([]string{
		"Edit main.go to customize your application",
		"Add more components from the GoFlow library",
		"Style your app with custom CSS",
		"Build something amazing with Go and WebAssembly!",
	}).SetClass("feature-list")
	features.AddChild(featureList)

	// Links section
	links := component.NewContainer().SetClass("links-section")
	linksTitle := component.NewHeading("Learn More", 3).SetClass("links-title")
	
	// Create interactive buttons
	docsBtn := component.NewButton("📚 Documentation", func() {
		fmt.Println("Opening GoFlow documentation...")
		// In a real app, this would open the docs
	}).SetClass("btn btn-primary")

	githubBtn := component.NewButton("⭐ GitHub", func() {
		fmt.Println("Opening GoFlow GitHub...")
		// In a real app, this would open GitHub
	}).SetClass("btn btn-secondary")

	websiteBtn := component.NewButton("🌐 GoFlow Website", func() {
		fmt.Println("Opening GoFlow website...")
		// In a real app, this would open the website
	}).SetClass("btn btn-accent")

	links.AddChild(linksTitle)
	links.AddChild(docsBtn)
	links.AddChild(githubBtn)
	links.AddChild(websiteBtn)

	// Footer
	footer := component.NewContainer().SetClass("footer-section")
	footerText := component.NewParagraph("Built with ❤️ using GoFlow - The Go WebAssembly Framework").SetClass("footer-text")
	footer.AddChild(footerText)

	// Assemble the page
	page.AddChild(header)
	page.AddChild(features)
	page.AddChild(links)
	page.AddChild(footer)

	return page
}
`

const indexHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoFlow App</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        /* GoFlow Brand Colors */
        :root {
            --primary-blue: #2196F3;
            --primary-orange: #FF9800;
            --gradient-blue: #42A5F5;
            --gradient-orange: #FFB74D;
            --text-primary: #2C3E50;
            --text-secondary: #546E7A;
            --light-gray: #F5F7FA;
            --white: #FFFFFF;
            --shadow: rgba(0, 0, 0, 0.1);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, var(--primary-blue) 0%, var(--primary-orange) 100%);
            min-height: 100vh;
            color: var(--text-primary);
            line-height: 1.6;
        }

        .loading {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100vh;
            color: white;
            text-align: center;
        }

        .loading-spinner {
            width: 50px;
            height: 50px;
            border: 4px solid rgba(255, 255, 255, 0.3);
            border-top: 4px solid white;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-bottom: 20px;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }

        /* Welcome Page Styles */
        .welcome-page {
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
        }

        .header-section {
            text-align: center;
            margin-bottom: 3rem;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(20px);
            padding: 3rem;
            border-radius: 20px;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .logo {
            font-size: 4rem;
            margin-bottom: 1rem;
            display: block;
            animation: rotate 10s linear infinite;
        }

        @keyframes rotate {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
        }

        .main-title {
            font-size: 3rem;
            font-weight: 700;
            margin-bottom: 1rem;
            color: white;
            text-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
        }

        .subtitle {
            font-size: 1.2rem;
            color: rgba(255, 255, 255, 0.9);
            margin-bottom: 0;
        }

        .features-section, .links-section {
            background: white;
            padding: 2.5rem;
            border-radius: 20px;
            margin-bottom: 2rem;
            box-shadow: 0 10px 40px var(--shadow);
        }

        .section-title, .links-title {
            color: var(--text-primary);
            margin-bottom: 1.5rem;
            font-weight: 600;
        }

        .section-title {
            font-size: 2rem;
            text-align: center;
        }

        .links-title {
            font-size: 1.5rem;
            text-align: center;
        }

        .feature-list {
            list-style: none;
            padding: 0;
        }

        .feature-list li {
            padding: 1rem 0;
            border-bottom: 1px solid var(--light-gray);
            color: var(--text-secondary);
            position: relative;
            padding-left: 2rem;
        }

        .feature-list li:before {
            content: '✨';
            position: absolute;
            left: 0;
            top: 1rem;
        }

        .feature-list li:last-child {
            border-bottom: none;
        }

        .links-section {
            text-align: center;
        }

        .btn {
            display: inline-block;
            padding: 12px 24px;
            margin: 0.5rem;
            border: none;
            border-radius: 12px;
            font-weight: 600;
            font-size: 1rem;
            cursor: pointer;
            transition: all 0.3s ease;
            text-decoration: none;
            font-family: inherit;
        }

        .btn-primary {
            background: linear-gradient(45deg, var(--primary-blue), var(--gradient-blue));
            color: white;
            box-shadow: 0 4px 15px rgba(33, 150, 243, 0.3);
        }

        .btn-secondary {
            background: var(--light-gray);
            color: var(--text-primary);
            border: 2px solid var(--primary-blue);
        }

        .btn-accent {
            background: linear-gradient(45deg, var(--primary-orange), var(--gradient-orange));
            color: white;
            box-shadow: 0 4px 15px rgba(255, 152, 0, 0.3);
        }

        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px var(--shadow);
        }

        .footer-section {
            text-align: center;
            padding: 2rem;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(20px);
            border-radius: 20px;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .footer-text {
            color: rgba(255, 255, 255, 0.9);
            font-size: 0.9rem;
            margin: 0;
        }

        /* Responsive Design */
        @media (max-width: 768px) {
            .welcome-page {
                padding: 1rem;
            }

            .header-section {
                padding: 2rem;
            }

            .main-title {
                font-size: 2.2rem;
            }

            .logo {
                font-size: 3rem;
            }

            .btn {
                display: block;
                margin: 0.5rem 0;
            }
        }

        .error {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100vh;
            color: white;
            text-align: center;
        }
    </style>
</head>
<body>
    <div id="app">
        <div class="loading">
            <div class="loading-spinner"></div>
            <h2>Loading GoFlow App...</h2>
            <p>Initializing WebAssembly...</p>
        </div>
    </div>

    <script src="wasm_exec.js"></script>
    <script>
        if (!WebAssembly.instantiateStreaming) {
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
            document.getElementById('app').innerHTML = 
                '<div class="error"><h2>⚠️ Failed to Load</h2><p>Check the browser console for details.</p></div>';
        });
    </script>
</body>
</html>
`

func goModTemplate(projectName string) string {
	return fmt.Sprintf(`module %s

go 1.25.1

require github.com/AureClai/goflow v0.1.0
`, projectName)
}

func readmeTemplate(projectName string) string {
	return fmt.Sprintf(`# %s

A GoFlow WebAssembly application generated by the GoFlow CLI.

## 🚀 Getting Started

### Prerequisites
- Go 1.25.1 or later
- Modern web browser with WebAssembly support

### Development

1. **Build the application:**
`+"   ```bash\n   goflow build\n   ```"+`

2. **Start the development server:**
`+"   ```bash\n   goflow dev\n   ```"+`

3. **Open your browser:**
   Visit http://localhost:8080 to see your app!

## 📁 Project Structure

- `+"`main.go`"+` - Your main application code
- `+"`index.html`"+` - HTML template with styling
- `+"`go.mod`"+` - Go module configuration
- `+"`app.wasm`"+` - Compiled WebAssembly binary (generated)
- `+"`wasm_exec.js`"+` - Go WebAssembly runtime (generated)

## 📚 Learn More

- [GoFlow Documentation](https://github.com/AureClai/goflow)
- [GoFlow Examples](https://github.com/AureClai/goflow/tree/main/examples)
- [WebAssembly with Go](https://golang.org/pkg/syscall/js/)

## 🛠️ Next Steps

1. Edit `+"`main.go`"+` to customize your application
2. Add more GoFlow components from the library
3. Style your app with custom CSS
4. Build something amazing with Go and WebAssembly!

---

Built with ❤️ using [GoFlow](https://github.com/AureClai/goflow) - The Go WebAssembly Framework
`, projectName)
}

const gitignoreTemplate = `
# Compiled Wasm file
app.wasm

# JS glue file
wasm_exec.js
`
