# Bittorrent Implementation: Final Project Writeup

## Introduction: What were your overall project goals? What (briefly) did you achieve?
Our goal was to implement a simple version of Bittorrent, the protocol used for initiating and managing peer-to-peer file transfers. We achieved a basic implementation of the protocol, allowing peers to connect to a tracker server, request a file using a torrent file that the tracker server then finds seeders for, and download it from other peers provided by the tracker server. We also added functionality for peers to upload a file, which meant creating the torrent file and registering as a seeder with the tracker server.

## Design/Implementation: What did you build, and how does it work? For this part, give an overview of the major components of your system design and how they work, similar what you might write in a readme.

Our system consists of two main components: the tracker server and the peer client.

### Tracker Server

The tracker server acts as a central database that keeps track of all the peers in the network and the torrent files they are sharing (crucially, not the files themselves). Its primary responsibilities include:

- **Registration of Seeders**: When a peer wants to share a file, it creates a torrent file containing metadata about the file and registers the info_hash with the tracker server and its peer_id as a seeder. The tracker updates its list of seeders for that file to include the new seeder, usually as the first seeder since its probably a new file.

- **Managing Peer Information**: The tracker maintains a list of active peers and the files they are sharing. It stores information such as peer IDs, IP addresses, and port numbers. It also tracks whether it is a seeder and when it last announced itself. That way, if a seeder has been idle more than a certain amount of time, the tracker can remove it from the list of seeders, which helps make sure that when the peer initiates the download protocol with that seeder, it is actually present.

- **Facilitating Peer Discovery**: When a peer wants to download a file, it contacts the tracker server to obtain a list of peers (seeders and leechers) who have the desired file. The tracker responds with the necessary information for the peer to establish direct connections.

- **Handling Announce Requests**: Peers periodically send announce requests to the tracker to update their status. The tracker processes these requests to keep its data current, iterating through the list of peers for a file and updating the last_announced field. repeats are allowed, since they will be removed eventually after a certain amount of time and there may be repeats in the returned list, but the peer should be able to iterate past any peers not seeding when they are contacted. 

- **Concurrency**: The server is equipped with mutexes for the maps to ensure thread safety when multiple peers are accessing the tracker server simultaneously.

- **Web Implementation**: The tracker server is implemented as a simple HTTP server that listens for incoming requests from peers. It uses a map to store information about the files and peers in the network (keeping a peer list for every info_hash, which is a sha1 representation of a file using its information). We run the tracker server on an Azure VM, which allows it to be accessible to peers on the internet.

### Peer Client

The peer client is responsible for downloading and uploading files. Its key functionalities include:

- **Torrent File Parsing**: The client can read torrent files to extract metadata such as file name, file size, and the hash of the file contents. Additionally, it can parse a file and create a torrent file from it. 

- **Connecting to the Tracker**: To download a file, the client contacts the tracker server using the information in the torrent file to get a list of available peers. From here, it iterates through the list, downloading as much as possible from each (in some cases, if the peer has disconnected, it will just move onto the next one), keeping track of which pieces of the file it has downloaded using a bitfield.

- **Communication protocol** We implemented a custom protocol for communication between peers and between peers and the tracker server (a simplified version of Bittorrent that doesn't include the full set of messages like choking and has a simplified and therefore not cross-compatible version of the info_hashes, among other small changes/non-inclusions). The tracker-peer connection is managed through HTTP requests while the peer-peer connection is managed through custom protocol messages sent through TCP.

- **Establishing Peer Connections**: Using the list from the tracker, the client establishes direct TCP connections with other peers to download file pieces. There is a handshake protocol that is used to communicate the info_hash and peer_id between peers, followed by the bitfield of which pieces are needed.

- **File Piece Management**: Files are divided into fixed-size pieces. The client keeps track of which pieces it has and which ones it needs, requesting specific pieces from peers using the bitfield. The leecher iterates through all of the pieces, requesting one piece at a time (checking if it already has it, for the case that this is a repeat calculation). Unlike the regular bittorrent, it is done in order of pieces; however, the functionality is close to what it would be if it was random order (or rarest-first order, which is how bittorrent is usually implemented) The piece index is included so the leecher can put the file together in the correct order. 

- **Uploading to Other Peers**: After finishing, a leecher announces to the server that it is done and the server adds the leecher as a now seeder for the file. 

- **Data Verification**: Each piece downloaded is verified using hashing to ensure data integrity before being written to disk.

- **Frontend:** We used Wails, a package that allows for the creation of desktop applications using Go and JavaScript, so we could build a local GUI for the peer client. The frontend allows the user to select a torrent file and start downloading the associated file, as well as upload a file and create a corresponding torrent. The backend here is in Go, which handles the actual downloading and uploading of files.


## Discussion/Results: Describe any results you have, what you have learned, and any challenges you faced along the way. For this part, please include any relevant logs/screenshots of your program operating (and/or reference your demo video).

Throughout the development of our Bittorrent implementation, we encountered several challenges and learning opportunities.

### Results Achieved

- **Port Forwarding**: Unfortunately, we couldn't get it working with remote IPs because of NAT, so automatic port forwarding would be great to try. Right now, we have our tracker server running remotely, but tracking local IPs and ports, but it would only require sending the remote IP address (as the client) and then implementing port forwarding in the client to make it work. We considered including a manual port forwarding version of this in the demo video, but since IP and TCP were done locally, we were content with leaving it local for now, wanting to implement other things first.  

- **Successful File Transfers**: We were able to demonstrate the complete cycle of uploading and downloading files between peers using our tracker server.

- **Peer Discovery Mechanism**: Implemented a functioning system where peers could discover others sharing the same file and establish direct connections.

- **Data Integrity Assurance**: By using hashing for each piece, we ensured that the files downloaded were accurate and free from corruption.

### Challenges Faced

- **Network Communication Complexity**: Handling low-level socket programming and ensuring reliable data transmission required extensive debugging and testing.

- **Concurrency Management**: Managing multiple threads for peer connections introduced complexity, including race conditions and synchronization issues.

- **Protocol Design**: Creating an efficient and reliable communication protocol between peers and the tracker was non-trivial and required careful consideration.

- **Error Handling**: Ensuring the system could gracefully handle unexpected disconnects, timeouts, and corrupted data was a significant challenge.

### Lessons Learned

- **Importance of Protocol Standards**: Adhering to well-defined protocols is crucial in network communication to ensure interoperability and reliability.

- **Robust Error Handling**: Anticipating and handling potential errors improves the resilience of the application.

- **Scalability Considerations**: Designing the system with scalability in mind is essential, especially for peer-to-peer networks where the number of peers can grow rapidly.

- **Concurrency Best Practices**: Effectively utilizing concurrency requires careful design to avoid common pitfalls associated with multi-threaded applications.


## Conclusions/Future work: Overall, what have you learned? How did you feel about this project overall? If you could keep working on this project, what would you do next?

### Overall Experience

This project provided invaluable hands-on experience with network programming and peer-to-peer systems. We deepened our understanding of the Bittorrent protocol and the complexities involved in distributed file sharing.

### Potential Improvements

If we were to continue working on this project, we would consider implementing the following enhancements:

- **Selective Piece Downloading**: Implement more advanced piece selection strategies, such as rarest-first, to optimize download efficiency. Also, implement the ability to download from multiple seeders at once.

- **Choke/Unchoke Mechanism**: Introduce choking algorithms to manage bandwidth and improve fairness among peers.

- **DHT Integration**: Incorporate Distributed Hash Tables to eliminate the reliance on a central tracker, increasing network robustness.

- **Encryption and Security**: Add encryption to peer communications to enhance security and prevent eavesdropping or tampering.

- **Full Peer Protocol Support**: Implement the full Bittorrent protocol, including the full set of messages, especially between peers.

- **Error Correction Protocols**: Implement mechanisms to recover from corrupted data beyond simple retransmission, such as using error-correcting codes.

- **Increased Anonymity**: Enhance privacy by implementing mechanisms to obfuscate peer identities and file transfers, especially from ISPs.

### Final Thoughts

This was an interesting project that allowed us to explore a fascinating protocol. Beyond all of the cool technical details, the ethical backdrop to it, with many Bittorrent implementations being limited by ISPs or outright banned, was also interesting to consider. It's interesting how much power a simple protocol can have over disrupting the distribution of software. There are both great usages for a protocol like this, for distributing censored material in an autocratic government or for distributing large files that would be expensive to host on a server, and also illegal and potentially harmful usages, like distributing copyrighted material without permission. 

We are proud of what we achieved and inspired to continue exploring the field of distributed networking and peer-to-peer technologies.
