# Off-Sync.com Platform Proxy Application

The application project holds the commands and queries from which the Platform Proxy is build, as well as all required external interfaces.

This application distinguishes the following aggregate roots:
* Services;
* Frontends.

## Services

Services provide functionality by exposing one or more backend servers.

### Get Services Query

The Get Services Query returns the list of currently configured backend services.

### Start Services Watcher Command

The Start Services Watcher Command is used to start a watcher on changes in the services configuration. The watcher is provided with a channel that can be used to push these changes back to the application.

## Frontends

Frontends define the way in which Services are exposed on the Platform Proxy.

### Get Frontends Query

The Get Frontends Query returns the list of currently configured frontends.

### Start Frontends Watcher Command

The Start Frontends Watcher Command is used to start a watcher on changes in the frontends configuration. The watcher is provided with a channel that can be used to push these changes back to the application.
