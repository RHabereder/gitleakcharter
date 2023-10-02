# Introduction 
This little tool is going to build several charts to illustrate the amount of leaks of which type were found in a git-repository
The only prerequisite is a json report that was generated via Gitleaks 

# Getting Started
1.	Install Gitleaks
2.	Create a report (Example: `gitleaks detect -c ../gitleaks.toml . -f json -r report.json`) 
3.  Run the Tool (`go run main.go -f path/To/Reports/Directory/ -o report.html`)

# Contributors
* Raphael Habereder