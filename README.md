# D7024E Luleå University of Technology, Kademlia DHT implemented in Go

## Introduction

This lab is part of the major examination in a **mobile and distributed computing** course. To accommodate the course's time frame, some simplifications have been made in this exercise. The lab focuses on implementing a Distributed Hash Table (DHT) using the Kademlia algorithm in Go, which is widely used in decentralized and peer-to-peer networks. The exercise, along with the mandatory and qualifying objectives, is detailed in the `Kademlia-lab\pdf\D7024E_Kademlia_lab.pdf`.

## Requirements

To run the system, only Docker is required. Ensure that Docker is properly installed and running on your machine before proceeding with the following steps.

## Running

1. Clone the repository to your local machine using the following command:

    ```bash
    git clone <repository-url>
    ```

2. Navigate to the Kademlia folder within the repository using your terminal:

    ```bash
    cd path/to/Kademlia
    ```

3. Start the system by running the Docker Compose command:

    ```bash
    docker compose up --build
    ```

If you are on Windows, there might be issues related to line ending encodings, as Windows uses `CRLF` while Unix-based systems use `LF`. If such issues arise during the build process, they will be highlighted in the raised errors. You will need to change the line endings of the specified files from `CRLF` to `LF` before rerunning the system. This can typically be resolved by using text editors like Visual Studio Code or Notepad++ to convert the line endings.
