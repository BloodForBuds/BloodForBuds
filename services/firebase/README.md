# Firebase Authentication Emulator

Development uses the Firebase Authentication Emulator with the isolated
`demo-bloodforbuds` project. It does not use staging or production credentials,
and it does not send real email. Email sign-in links are available from the
emulator API and container logs.

The emulator is attached only to the private Compose network. Caddy proxies the
Firebase Auth API paths needed by the browser, so the emulator does not publish
a host port.

Staging and production use their real Firebase projects and do not run this
container.
