Feature: Authentication and Authorization

    @AUTH-C-01
    @Backend
    Scenario: Authentication

        All interface functions must have authentication

        Given any gRPC interface
        When an unauthorized request arrives
        Then the request must be rejected


    @AUTH-C-02
    @Backend
    Scenario: Standards

        To be compliant with standards and best-practices OAuth 2.0 in combination with JWT tokens should be used for all components and interfaces that need authentication.

        Given any gRPC interface
        When a valid JWT arrives
        Then the request must be fulfilled

    @AUTH-C-03
    @Frontend
    Scenario: Deprecated OAuth 2.0 Flows

        The components must avoid authentication configurations that are considered to be deprecated, e.g., OAuth 2.0 implicit flow.

        Given the dashboard
        When the user logins
        Then no implicit flow must take place

    @AUTH-C-04
    @Backend
    Scenario: Sensible Security Defaults

        JWT tokens must have sensible security defaults, e.g., with regards to expiration date

        Given any gRPC interface
        When a token with expiration older than 24h arrives
        Then the request must be rejected
