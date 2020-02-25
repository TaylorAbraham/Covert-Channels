import React from 'react';
import { mount } from 'enzyme';
import { act } from 'react-dom/test-utils';
import MessagingScreen from './MessagingScreen';

const covertMessages = [];
const sendMessage = () => {};

describe('Messaging screen', () => {
  it('handles input ', () => {
    const setTextToSend = jest.fn();
    let textToSend = '';
    let msgScreen;
    act(() => {
      msgScreen = mount(<MessagingScreen
        textToSend={textToSend}
        setTextToSend={setTextToSend}
        covertMessages={covertMessages}
        sendMessage={sendMessage}
      />);
    });
    expect(msgScreen.find('.cc-messaging__send').at(0).props().disabled).toBe(true);
    act(() => {
      textToSend = 'a';
      msgScreen.find('.cc-messaging__input').at(0).simulate('change', { target: { value: 'a' } });
    });
    expect(setTextToSend).toHaveBeenCalled();
    expect(msgScreen.find('.cc-messaging__send').at(0).props().disabled).toBe(false);
    expect(msgScreen).toMatchSnapshot();
  });
});
