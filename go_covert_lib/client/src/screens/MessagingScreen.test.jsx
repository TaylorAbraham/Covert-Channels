import React from 'react';
import { mount } from 'enzyme';
import { act } from 'react-dom/test-utils';
import MessagingScreen from './MessagingScreen';

const covertMessages = [];
const sendMessage = () => {};

describe('Messaging sescreen', () => {
  it('handles input in the input field', () => {
    const textToSend = '';
    let msgScreen;
    const setTextToSend = jest.fn(val => msgScreen.setProps({ textToSend: val }));
    act(() => {
      msgScreen = mount(<MessagingScreen
        textToSend={textToSend}
        setTextToSend={setTextToSend}
        covertMessages={covertMessages}
        sendMessage={sendMessage}
      />);
    });
    let sendButtonIsDisabled = msgScreen.find('.cc-messaging__send').at(0).props().disabled;
    expect(sendButtonIsDisabled).toBe(true);
    act(() => {
      // Simulate typing a character
      msgScreen.find('.cc-messaging__input').at(0).simulate('change', { target: { value: 'a' } });
    });
    sendButtonIsDisabled = msgScreen.find('.cc-messaging__send').at(0).props().disabled;
    expect(setTextToSend).toHaveBeenCalled();
    expect(sendButtonIsDisabled).toBe(false);
    expect(msgScreen).toMatchSnapshot();
  });
});
