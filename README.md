Romeo
=====

Simple services control toolkit


## What is service?

In terms of toolkit, **service** is structure, able to start and stop. This requirements
are formalized in `oscar.Service` interface:

```go
type Service interface {
	Start(xray.Ray) error
	Stop(xray.Ray) error
}
```

Any component, that contains that method is valid service and can be handled by Romeo.

Each service can optionally provide it's name by `GetName() string` method. This name 
currenlty used only for logging purposes.

## Run levels

`RunLevel` is a service startup priority. RunLevel is simple wrapper over `byte` and 
can be omitted during service configuration and Romeo will assign `RunLevelMain` for it.

If your service must be invoked in custom priority order, it must have method 
`GetRunLevel() RunLevel`. 

There are following major predefined runlevels:

| RunLevel | Byte code | Description |
| -------- | --------- | ----------- |
| RunLevelDB | 100 | To be used for database sources |
| RunLevelReloaders | 120 | Used by components, that on regular basis update themselves. |
| RunLevelAPIServer | 140 | Used by API service (RabbitMQ, for example) |
| RunLevelMain | 160 | Main application priority |

## Server

Romeo bundles with common server `server.Server`. You can register own services in it 
using `Server.Register(...Service)` method, and them start all of them using 
`Server.Start` method. Key notes:

1. Server starts all services it contains
2. Startup process is devided into stages. Each runlevel has own stage.
3. All services withing one stage (runlevel) starts in parallel. 
   Stages stats consequentially.
4. If at least one service within stage fails, all stage became failed and
   startup process stops.
5. Startup uses runlevels in ascending order. 
6. Shutdown stops all services in descending order.

## Service helpers

### `services.Container`

`Container` is service, that contains within it multiple other services. It uses slice
`[]romeo.Service` internally and can be created using type casting:

```go
cont := service.Container([]romeo.Service{one, two, three})
```

### 
