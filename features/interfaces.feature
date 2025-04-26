Feature: Interfaces

    @INTFC-C-01
    @Backend
    Scenario: RPC between components

        While in general, the components must follow the Gaia-X approach for REST-based APIs,
        REST may not suffice for certain scenarios when transmitting data within the service.
        Therefore, when needed (see INTFC-C-02 and INTFC-C-03), other RPC protocols, such as gRPC,
        must be chosen for transmitting data between sub-components within the CAM.

        Given any component
        When RPC mechanisms are needed
        Then gRPC must be used

    @INTFC-C-02
    @Backend
    Scenario: RPC for events

        Functionalities within an interface that relate to triggers and events cannot be sufficiently
        represented in REST, whereas an RPC call would allow for easy interaction and also subscribing
        to certain messages. Therefore, those functionalities must be implemented in the RPC mechanism
        chosen in INTFC-C-01.

        Given any component
        When the component has events
        Then an RPC call must be used

    @INTFC-C-03
    @Backend
    Scenario: RPC for streaming

        Functionalities within an interface that relate to triggers and events cannot be sufficiently
        represented in REST, whereas an RPC call would allow for easy interaction and also subscribing
        to certain messages. Therefore, those functionalities must be implemented in the RPC mechanism
        chosen in INTFC-C-01.

        Given any component
        When the component has events
        Then an RPC call must be used

    @INTFC-C-04
    @Backend
    Scenario: gRPC Gateway

        The choice of a non-REST interface should at least provide capabilities to convert or expose
        them as REST interfaces. For example, through projects such as the gRPC Gateway3, certain gRPC
        interfaces can easily be exposed through a REST API, for those interfaces, where it makes sense,
        i.e., the dashboard.

        Given any interface
        When the RPC call must be exposed as REST
        Then a gRPC gateway must be used
