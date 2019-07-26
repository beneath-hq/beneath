# Beneath-models

The purpose of the Models is to retrieve parse and post basic data to the Beneath gateway. The first Model handles data from the Ethereum network via any client or service exposing the Ethereum HTTP JSON-RPC API.

## Setting up the development environment

The Models are written in Python 3 and uses several liberaries and tools for testing and linting. The following steps assume you are on MacOS, but other popular platforms should have quite similar steps.

### 1. Installing Python

(reference: https://docs.python-guide.org/starting/install3/osx/)

1.1 Install XCode with `xcode-select --install`

1.2 Install Homebrew with `/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`

1.2.1 Confirm that Homebrew is installed correctly with `brew doctor`

1.3 Install Python 3 with `brew install python3`

1.3.1 Confirm that Python 3 is installed correctly with `python3 --version`

### 2. Clone the repository

Assuming you have already installed git. If you have not already done so, clone beneath-core by running `git clone https://github.com/beneathcrypto/beneath-core.git`

### 3. Install required Python modules

3.1 In the cloned repository, change directory to `/beneath-core/beneath-models`

3.2 Initialize the Python virtual environment and install the required modules by running `./init-virtual-env.sh`

### 4. Run the tests

To run the tests, simply run `pytest` (read more about pytest options here: https://pytest.org)

### 5. Run the Model

5.1 Activate the virtual environment by running `source .venv/bin/activate`

5.2 Go to the model directory `cd eth-block-numbers`

5.2 Run the Model with `python3 fetch_and_publish_blocks.py`

5.3 When you are done running the Model, exit the virtual environmen by simply running `deactivate`

### 6. Use the virtual environment in VS Code

For VS Code to detect the virtual environment, it (.venv) must be located in the VS Code workspace root. If in doubt, make sure that "beneath-models" is the diretory you opened in VS Code.

To use the virtual environment in VS Code, press `cmd+shift+p`, search and select `Python: Select Interpreter`, and select the Python 3 executable with the '.venv' virtual environment - something like: `Python 3.7.3 64-bit ('.venv': virtualenv)`.