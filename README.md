Project Abandoned
=================
This is a code dump of what I had hoped to turn into a business, but walked away from.  No support or maintenance will be provided.


Summit Route End Point Protection (SREPP) - Server code
===============================================

Proof-of-concept code. Still needs lots of work to make it a complete product.

Architecture
------------
The server is composed of 4 parts:

- frontend: ReactJS javascript project to display a UI for the customer.  Some parts of this are simply mocks with no functionality or fake data.
- WebServer: Go code to provide the frontend pieces and APIs to collect data from the database
- CallbackServer: Go code for APIs the agents to communicate with.  Agents beacon data which is written to the database, and potentially receiving tasking (such as collect an executable).  Copies of executables are also sent back to the callbackserver which writes them to disk and creates tasks for workers to analyze.
- worker: Python code for analyzing the PE file signature on any executables.


Running
-------
The project was built to both be runnable on AWS or locally in a VM, so it was not tied to AWS and could deployed on-prem if needed.  It was meant to be horizonally scalable, but this was barely tested.  Deployed on Debian 7.7 x86_64. It uses Postgress for it's database and RabbitMQ for it's queueing.

The WebServer runs on port 8000 but is connected to via an nginx proxy for load-balancing and SSL termination that receives traffic on 443.
Likewise the Callback server runs on 8080, but has nginx in front of it receiving traffic on 8443. 

- Create and start the Postgress database.
- Start RabbitMQ.
- Rename this project `qdserver`
- Build the frontend (run `gulp` in ./frontend)
- In WebServer, run `go run server.go`
- In CallbackServer, run `go run server.go`