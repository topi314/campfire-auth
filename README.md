# campfire-auth

Campfire Auth is a standalone authentication service for Campfire, providing a way for users to log in without their credentials.

## How it works

A user is redirected to Campfire Auth, where they are prompted to send a short code into a specific channel in a campfire group.
Upon seeing the code, Campfire Auth verifies the user and redirects them back to the original service with a code that can be exchanged for their user object.

The server requires a client secret to be used when exchanging the code for a user object, ensuring that only authorized services can access user data.

Additionally, the server provides an endpoint to get a user by their ID and an endpoint to search for users by their username.
Both endpoints require the client secret to be provided too.

## Usage

If you want to use Campfire Auth in your application, you can either self-host it or use the hosted version at https://auth.cmpf-tools.de.
To use the hosted version contact me on one of the platforms listed below.

## Docs

The full API documentation is available at https://auth.cmpf-tools.de/api/docs

## License

This project is licensed under the [Apache License 2.0](LICENSE).

## Contributing

Contributions are welcome, but for bigger changes, please open an issue first to discuss what you would like to change.

## Contact

- [Discord](https://discord.gg/sD3ABd5)
- [Matrix](https://matrix.to/#/@topi:topi.wtf)
- [Twitter](https://twitter.com/topi314)
- [Email](mailto:hi@topi.wtf)
