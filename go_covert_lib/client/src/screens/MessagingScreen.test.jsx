import React from 'react';
import { mount } from 'enzyme';
import MessagingScreen from './MessagingScreen';

describe('Messaging screen', () => {
  const covertMessages = [];
  const sendMessage = jest.fn();
  const textToSend = '';
  let msgScreen;
  const setTextToSend = jest.fn(val => msgScreen.setProps({ textToSend: val }));

  beforeEach(() => {
    msgScreen = mount(<MessagingScreen
      textToSend={textToSend}
      setTextToSend={setTextToSend}
      covertMessages={covertMessages}
      sendMessage={sendMessage}
    />);
  });

  afterEach(() => {
    msgScreen.unmount();
  });

  it('handles input in the input field', () => {
    let sendButtonIsDisabled = msgScreen.find('.cc-messaging__send').at(0).props().disabled;
    expect(sendButtonIsDisabled).toBe(true);
    // Simulate typing a character
    msgScreen.find('.cc-messaging__input').at(0).simulate('change', { target: { value: 'a' } });
    sendButtonIsDisabled = msgScreen.find('.cc-messaging__send').at(0).props().disabled;
    expect(setTextToSend).toHaveBeenCalled();
    expect(sendButtonIsDisabled).toBe(false);
    expect(msgScreen).toMatchSnapshot();
  });

  it('handles input submission', () => {
    msgScreen.find('.cc-messaging__input').at(0).simulate('change', { target: { value: 'a' } });
    const sendButton = msgScreen.find('.cc-messaging__send').at(0);
    sendButton.simulate('click');
    expect(sendMessage).toHaveBeenCalled();
  });

  it('displays incoming messages', () => {
    const msg1 = 'Coca-Cola';
    const msg2 = 'Never forget';
    msgScreen.setProps({ covertMessages: [msg1, msg2] });
    const outputField = msgScreen.find('.cc-messaging__output').at(0);
    expect(outputField.props().value).toEqual(`${msg1}\n${msg2}`);
    expect(msgScreen).toMatchSnapshot();
  });
});
