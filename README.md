# Introduction 
This little tool is going to build several charts to illustrate the amount of leaks of which type were found in a git-repository
The only prerequisite is a json report that was generated via Gitleaks 

# Getting Started
TODO: Guide users through getting your code up and running on their own system. In this section you can talk about:
1.	Install Gitleaks
2.	Create a report (Example: `gitleaks detect -c ../gitleaks.toml . -f json -r report.json`) 
3.	Set the Env-Var for the input-file (`LC_INPUT_FILE=./report.json`)
4.	Set the Env-Var for the Pie-Chart Output (`LC_BAR_OUTPUT_FILE=pie.png`)
5.	Set the Env-Var for the Bar-Chart Output (`LC_BAR_OUTPUT_FILE=bar.html`)
6.  Run the Tool (`go run main.go`)