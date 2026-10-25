## Running the program

When running the program you have the option to give 2 flags on the clients. The first flag -name is the name of the given client, if not specified
it will default to the value "NoNameUser". 
The second flag which is addr, is the address of the server, the default value is localhost:9000, this is also the adress the server runs on and should not be
changed in the client without being aware the connection will not work unless the value is changed equally in the server.

To run the program, run "go run ." in the server folder. In the client folder you can run "go run ." to run a client. To specify a name you run "go run . -name Example"
This will create a client instance with the name "Example"

If 2 clients connect without specifying a name this will create an error. This is a problem that, if the chitchat should expand, should be the first to be fixed.
Though after consulting with a TA this was allowed for this task.