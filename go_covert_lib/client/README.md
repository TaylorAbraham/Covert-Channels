# Covert Client

## Setup

The client is served by the server, but dependencies must be installed and the client must be built first. You may do both by running
```
npm install
npm run build
```

If you are making changes to the frontend, rather than re-building the frontend every time, you may use the client watch script to automatically trigger a build each time changes are made. First run the one-time setup by allowing execution and installing inotify-tools:
```
chmod +x client_watch.sh
sudo apt update && sudo apt install inotify-tools
```

Now the build script can be run in a terminal via
```
./client_watch.sh
```

## Making changes

For styling, refer to [the bootstrap style guides](https://getbootstrap.com/docs/4.0/utilities/position/). There are many pages in the stlye guide available under "Utilities" on the left-hand sidebar

Make sure to fix any linting errors before committing.
