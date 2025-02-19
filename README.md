WIP for a distributed monolith.

Overview

Gonolith is designed to find the right balance between microservice and monolith architectures. It acknowledges a fundamental reality: most microservices ultimately become distributed monoliths due to their service interdependencies, where all components must be operational for the system to function. Rather than fighting this tendency, Gonolith embraces it as a core design principle, allowing for better chaos management.

Service Installation API:

Each node in the distributed monolith exposes an installService() API that accepts binary executables. This enables Gonolith to function as a hub where services can:

    Declare their presence and capabilities
    Connect with other services via RPC
    Potentially reduce unnecessary network hop overhead

Flexible Service Distribution:

Gonolith doesn't mandate that every node must contain every service. This flexible approach offers several advantages:

    Enables independent service scaling
    Maintains the cost-efficiency benefits of microservice architecture
    Allows for resource optimization

Node Configuration:

The system operates through sophisticated configuration management:

    Supports both individual node and node group configurations
    Allows nodes to host multiple services based on capacity
    Enables minimal service distribution when performance requires it

TODO:
- [ ] Gossip protocal
- [ ] Automatic updates via config (binary stored on ~s3 and fetched via config url)
