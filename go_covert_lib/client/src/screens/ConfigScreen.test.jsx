/* eslint-disable no-use-before-define */
import React from 'react';
import { mount } from 'enzyme';
import { act } from 'react-dom/test-utils';
import ConfigScreen from './ConfigScreen';

const sampleProcessorList = {
  Caesar: {
    Shift: {
      Type: 'i8',
      Value: 49,
      Range: [
        -128,
        127,
      ],
      Display: {
        Description: 'The shift for the Caesar cypher',
        Name: 'Shift',
        Group: '',
        GroupToggle: false,
      },
    },
  },
  Checksum: {
    Polynomial: {
      Type: 'select',
      Value: 'IEEE',
      Range: [
        'IEEE',
        'Castagnoli',
        'Koopman',
      ],
      Display: {
        Description: 'The predefined polynomial to use for the crc32 checksum',
        Name: '',
        Group: '',
        GroupToggle: false,
      },
    },
  },
};

const sampleChannelList = {
  TcpSyn: {
    FriendIP: {
      Type: 'ipv4',
      Value: '127.0.0.1',
      Display: {
        Description: 'Your friend\'s IP address.',
        Name: 'Friend\'s IP',
        Group: 'IP Addresses',
        GroupToggle: false,
      },
    },
    OriginIP: {
      Type: 'ipv4',
      Value: '127.0.0.1',
      Display: {
        Description: 'Your IP address.',
        Name: 'Your IP',
        Group: 'IP Addresses',
        GroupToggle: false,
      },
    },
  },
  TcpHandshake: {},
};

// Polyfill which mocks document functions needed for Bootstrap components
global.document.createRange = () => ({
  setStart: () => {},
  setEnd: () => {},
  commonAncestorContainer: {
    nodeName: 'BODY',
    ownerDocument: document,
  },
});

describe('Config Screen', () => {
  let cfgScreen;
  const openChannel = jest.fn();
  const closeChannel = jest.fn();
  const config = {};
  const setConfig = jest.fn();
  const processorList = sampleProcessorList;
  const processors = [];
  const setProcessors = jest.fn();
  const channelList = sampleChannelList;
  const channel = {};
  const setChannel = jest.fn(() => cfgScreen.setProps({
    config: channelList.TcpSyn,
    channel: {
      value: 'TcpSyn',
      properties: channelList.TcpSyn,
    },
  }));
  const channelIsOpen = false;

  beforeEach(() => {
    cfgScreen = mount(<ConfigScreen
      openChannel={openChannel}
      closeChannel={closeChannel}
      config={config}
      setConfig={setConfig}
      processorList={processorList}
      processors={processors}
      setProcessors={setProcessors}
      channelList={channelList}
      channel={channel}
      setChannel={setChannel}
      channelIsOpen={channelIsOpen}
    />);
  });

  afterEach(() => {
    act(() => {
      cfgScreen.unmount();
    });
  });

  it('displays a configuration object correctly', async () => {
    expect(cfgScreen.debug()).toMatchSnapshot();
    await act(async () => {
      const chanSelect = cfgScreen.find('.cc-config__chan-select').at(0);
      chanSelect.simulate('click');
    });
    cfgScreen.update(); // Need to update because Bootstrap dropdowns are fancy and finicky
    const firstDropdownOption = cfgScreen.find('option').at(0);
    firstDropdownOption.simulate('click');
    expect(setChannel).toHaveBeenCalledTimes(1);
    expect(cfgScreen.debug()).toMatchSnapshot();
  });

  it('submits a configuration correctly', () => {
    setChannel();
    const submitButton = cfgScreen.find('.cc-config__submit').at(0);
    submitButton.simulate('click');
    expect(openChannel).toHaveBeenCalledTimes(1);
    expect(cfgScreen.debug()).toMatchSnapshot();
  });
});
