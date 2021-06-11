# RPC
This project implements an RPC server in Go using gRPC as a transport.

# Design
What follows is a design document outlining the technology choices, architecture, and implementation of the RPC client and server. 

## Authentication and Authorization
The RPC service includes mutual TLS authentication. The client and server will both have their own private keys as well as certificates signed by the same CA. The minimum TLS version will be set to TLS1.2. In order to protect against downgrade attacks in TLS1.2, an allowlist will be used for ciphers. Conveniently, the Go TLS package will return a list of all cipher suites it implements excluding suites with known security issues (tls.CipherSuites()). In order to restrict the client to using these only, PreferServerCipherSuites will be set to true in the server TLS configuration. If this were a production application, an endpoint would be included in the API to accept a certificate revocation list, allowing the administrator to 'ban' certain clients. 

For Authorization, a simple form of role based access control will be implemented: Each role will have its own CA. All CAs will be trusted by the server. Upon handling a remote procedure call, the server will first check the client certificate's certificate authority and compare it to its internal mapping of certificate authority to privilege level. The following roles will be implemented:
- **Read only:** this role can get job output, but not create or delete jobs. This is useful for automated tooling such as status indicators and job notifiers.
- **Read/Write access:** this role can get job output as well as create and delete jobs. This role is useful for administrators and any kind of deployment automation that would leverage this service.

![visual representation of the multiple client CA model](doc/client_cas.png)

## API
### Endpoints
The API exposed by the RPC service will contain the following endpoints
- **CreateJob**: receives a Job and returns a CreateJobOutput
- **DeleteJob**: receives a PID and returns a DeleteJobOutput
- **ListJobs**: returns a JobList
- **GetOutput**: receives a PID and returns a JobOutput

### Datatypes
- **Job**
    - Command (string):
    Command holds a valid shell command that is run on the underlying server shell.
    - PID (integer):
    PID holds an integer representing job ID. In this project it will correlate directly to the pid of the process running on the RPC server for a given job. In a production RPC system it is advisable to use a randomized PID.
    - CreatedTime (string):
    Timestamp that represents when the process was launched on the server
- **JobList**
    - Jobs ([]Job):
    A list of Job obejcts as defined above in this list
- **JobOutput**
    - Command (string):
    Holds the shell command as originally defined in the corresponding call to CreateJob
    - CurrentOutput (string):
    Current cummulative contents of stdout from Job
    - Complete (boolean):
    Boolean that represents whether the job is still running or if it has finished
- **DeleteJobOutput**
    - Job (Job):
    Job object as defined above in this list
    - Exit (integer):
    Exit code as returned from the underlying process
    - Interrupted (boolean):
    Boolean is set to true if underlying process was still running when DeleteJob was called
- **CreateJobOutput**
    - Job (Job):
    Created Job object as defined above in this list
    - Success (boolean):
    Boolean set to true if the job was launched successfully
    - Error (string):
    If success is set to False, an error can be found in Error
- **PID**
    Single integer representing the PID of the underlying process

### Data flow
Each handler will receive the endpoint input as well as a context object from which the client certificate and trust chain can be extracted. After the Authorization routine is applied to the CA certificate found in the client certificate trust chain, the handler will process the input and return the output.

![flowchart for CreateJob API call](doc/data_flow.png)

## Controller
The backend will leverage the CommandContext type from the os/exec package. This will provide convenient facilities for cancelling running jobs and garnering job output. Additionally, it will handle passing through environment variables for us.

A single datatype, JobCtl, will be defined to contain the following peices of data:
- The os/exec CommandContext instance
- The cancel function derived from teh context passed into the CommandContext instance
- The "Job" instance as defined in the API datatypes.
    
A sync.Map will be used to cache JobCtl objects. The JobCtl objects will be mapped to their PID. The sync.Map will take care of synchronization, reducing the likelihood of concerrency errors like race conditions. Methods will be defined to retrieve API Job objects, create new JobCtls, delete existing JobCtls, and fetch output from a JobCtl.

## Client
A command line client will be created from the compiled client library stub code.
A client docker image will be created which will contain the command line client as well as a specific client certificate and private key. The client docker image will also contain the server certificate. A unique docker image will be built for each client role (read only and read/write). All client image will use the same command line client application binary.

## Build environment
This project will be built and tested on a standard debian installation. A Makefile will be leveraged to automate building, testing, and deployment. The Makefile will create docker images to run the client and server code in. The docker images created by the Makefile will be based on Alpine linux. Alpine linux was chosen because it is a small, lightweight, and convenient solution. The Makefile will also provide a target that will run unit tests. 
