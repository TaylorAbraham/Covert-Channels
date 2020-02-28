import React from 'react';

const HelpScreen = (props) => {
  return (
    <div className="cc-help m-3">
      <h3>Overview</h3>
      <p>The purpose of this application is to be used as a research tool in the area of covert channel communication. Through the app, the user can create covert channels, configure them, and send covert messages across them. The user has the option to select from a variety of network protocols, encryption methods, and compression techniques to aid in the concealment of the message.</p>
      <h3>What is a Processor?</h3>
      <p>A processor is an entity that implements any operation on the transmitted or received messages. It is used to aid in the obfuscation or concealment of data for research-based purposes. The modifications currently supported are a variety of encryption and compression methods. A processor can be added by simply clicking the "Add Processor" button on the top of the configuration page and select a processor from the list. Multiple processors can be added at the same time to provide further obfuscation. Note: The processor(s) will be applied upon the opening of the covert channel.</p>
      <h3>Sending Messages to Another Client</h3>
      <p>To send messages to another client, ensure that you have a second instance of the server running on the same computer or a different one. Next, set the channel type by clicking the "Select a Channel" dropdown and clicking on the desired channel type. Note that all configuration options must be identical on the two clients for them to communicate properly. After that, swap the values of the "Friend's Port" and "Your Port" and click the "Open Channel" button at the very bottom of the page. Do the same with your client, but DO NOT switch port values. This will open up two complementary channels that can communicate with each other.</p>
      <p>Now navigate to the "Messaging" tab of each client. Here, try sending a message and it should be received on the other client. If this is the case, the application will have successfully sent a message to another client.</p>
      <h3>Other Information</h3>
      <p>For further information regarding each field within the application, tooltips are found nearby describing the use and purpose of the field.</p>
    </div>
  );
};

export default HelpScreen;
