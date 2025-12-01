# CapyChain Installation Guide

Welcome to the CapyChain installation guide. This document will walk you through the steps required to install and set up the CapyChain blockchain application on your system.

## Prerequisites

Before installing CapyChain, ensure that you have the following prerequisites:
- Git installed on your system
- Go programming language installed (version 1.25 or higher)
- Sqlite3 installed (for database support)

## Installation Steps

1. **Clone the Repository**
    Open your terminal and run the following command to clone the CapyChain repository:
    ```bash
    git clone https://github.com/joeCavZero/capychain.git
    ```
2. **Build the Application**
    Navigate to the cloned directory and build the application using Go:
    ```bash
    cd capychain
    go build
    ```
3. **Copy CapyChain Executable and interface.html**
    After building the application, copy the `capychain` executable and `interface.html` file to a separated directory where you want to run CapyChain:

    This step is important if you plan to run multiple instances of CapyChain, as each instance requires its own separate data directory.
    
4. **Run CapyChain**
    You can now run the CapyChain application using the following command:
    ```bash
    ./capychain <node_name> <port>
    ```
    Replace `<node_name>` with a unique name for your node and `<port>` with the port number you want the node to listen on (e.g., 8080).

    After running the command, CapyChain will start and create a `capychain.sqlite3` database file in the current directory to store blockchain data.

    It will also display the address and port where the web interface can be accessed.
5. **Access the Web Interface**
    Open your web browser and navigate to `http://<address>:<port>/interface` to access the CapyChain web interface. Replace `<address>` and `<port>` with the appropriate values given during the run command.
