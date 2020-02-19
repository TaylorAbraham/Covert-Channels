import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import Button from 'react-bootstrap/Button';
import Dropdown from 'react-bootstrap/Dropdown';

import IPInput from '../ui-components/IPInput';
import NumberInput from '../ui-components/NumberInput';
import Checkbox from '../ui-components/Checkbox';
import Select from '../ui-components/Select';

const ConfigScreen = (props) => {
  const {
    openChannel,
    closeChannel,
    config,
    setConfig,
    processorList,
    processors,
    setProcessors,
    channelList,
    channel,
    setChannel,
  } = props;

  return (
    <div className="m-2">
      <h2 className="m-1">Configuration</h2>
      <h3 className="m-1">Processors</h3>
      <Button
        variant="success"
        className="m-1 w-25"
        onClick={() => setProcessors(processors.concat({
          Type: null,
          Data: null,
        }))}
      >
        Add Processor
      </Button>
      {
        processors.map((processor, i) => (
          <div key={i.toString()}>
            <Dropdown className="m-1">
              <Dropdown.Toggle
                className="w-25"
                variant="outline-primary"
              >
                {processor.Type || 'Select a Processor'}
              </Dropdown.Toggle>
              <Dropdown.Menu className="w-25">
                {
                  Object.keys(processorList).map(p => (
                    <Dropdown.Item
                      as="option"
                      active={p === processor.Type}
                      onClick={(e) => {
                        setProcessors([
                          ...processors.slice(0, i),
                          {
                            Type: e.target.value,
                            Data: processorList,
                          },
                          ...processors.slice(i + 1, processors.length + 1),
                        ]);
                      }}
                      value={p}
                      key={p}
                    >
                      {p}
                    </Dropdown.Item>
                  ))
                }
              </Dropdown.Menu>
            </Dropdown>
            {processor.Data && Object.keys(processor.Data[processor.Type]).map((key) => {
              /**
               * EXTREME DANGER WARNING
               * The below code involves very convoluted spread operators to massage
               * the data to the format that GoLang expects it to be in.
               */
              const opt = processor.Data[processor.Type][key];
              const props = {
                key,
                label: opt.Display.Name,
                value: opt.Value,
                onChange: e => setProcessors([
                  ...processors.slice(0, i),
                  {
                    ...processor,
                    Data: {
                      ...processor.Data,
                      [processor.Type]: {
                        ...processor.Data[processor.Type],
                        [key]: {
                          ...opt,
                          Value: e.target.value,
                        },
                      },
                    },
                  },
                  ...processors.slice(i + 1, processors.length + 1),
                ]),
              };
              switch (opt.Type) {
                case 'ipv4':
                  return (
                    <IPInput {...props} />
                  );
                case 'i8':
                case 'u16':
                case 'u64':
                case 'exactu64':
                  return (
                    <NumberInput
                      {...props}
                      onChange={e => setProcessors([
                        ...processors.slice(0, i),
                        {
                          ...processor,
                          Data: {
                            ...processor.Data,
                            [processor.Type]: {
                              ...processor.Data[processor.Type],
                              [key]: {
                                ...opt,
                                Value: parseInt(e.target.value) || 0,
                              },
                            },
                          },
                        },
                        ...processors.slice(i + 1, processors.length + 1),
                      ])}
                    />
                  );
                case 'bool':
                  return (
                    <Checkbox
                      {...props}
                      onChange={e => setProcessors([
                        ...processors.slice(0, i),
                        {
                          ...processor,
                          Data: {
                            ...processor.Data,
                            [processor.Type]: {
                              ...processor.Data[processor.Type],
                              [key]: {
                                ...opt,
                                Value: e.target.checked,
                              },
                            },
                          },
                        },
                        ...processors.slice(i + 1, processors.length + 1),
                      ])}
                    />
                  );
                case 'select':
                  return (
                    <Select
                      {...props}
                      items={opt.Range}
                    />
                  );
                default:
                  return (<div key={key}>UNIMPLEMENTED</div>);
              }
            })}
          </div>
        ))
      }
      <h3 className="m-1">Channel</h3>
      <Dropdown className="m-1">
        <Dropdown.Toggle
          className="w-25"
          variant="outline-primary"
        >
          {channel.value || 'Select a Channel'}
        </Dropdown.Toggle>
        <Dropdown.Menu className="w-25">
          {
            Object.keys(channelList).map(chan => (
              <Dropdown.Item
                as="option"
                active={chan === channel.value}
                onClick={(e) => {
                  setChannel({
                    value: e.target.value,
                    properties: channelList[chan],
                  });
                  setConfig(channelList[chan]);
                }}
                value={chan}
                key={chan}
              >
                {chan}
              </Dropdown.Item>
            ))
          }
        </Dropdown.Menu>
      </Dropdown>
      {Object.keys(config).map((key) => {
        const opt = config[key];
        const props = {
          key,
          label: opt.Display.Name,
          value: opt.Value,
          onChange: e => setConfig({
            ...config,
            [key]: {
              ...config[key],
              Value: e.target.value,
            },
          }),
        };
        switch (opt.Type) {
          case 'ipv4':
            return (
              <IPInput {...props} />
            );
          case 'i8':
          case 'u16':
          case 'u64':
          case 'exactu64':
            return (
              <NumberInput
                {...props}
                onChange={e => setConfig({
                  ...config,
                  [key]: {
                    ...config[key],
                    Value: parseInt(e.target.value) || 0,
                  },
                })}
              />
            );
          case 'bool':
            return (
              <Checkbox
                {...props}
                onChange={e => setConfig({
                  ...config,
                  [key]: {
                    ...config[key],
                    Value: e.target.checked,
                  },
                })}
              />
            );
          case 'select':
            return (
              <Select
                {...props}
                items={opt.Range}
              />
            );
          default:
            return (<div key={key}>UNIMPLEMENTED</div>);
        }
      })}
      <Button variant="success" onClick={openChannel} className="m-1 w-25">Open Covert Channel</Button>
      <Button variant="danger" onClick={closeChannel} className="m-1 w-25">Close Covert Channel</Button>
    </div>
  );
};

ConfigScreen.propTypes = {
  openChannel: PropTypes.func.isRequired,
  closeChannel: PropTypes.func.isRequired,
  config: PropTypes.object.isRequired,
  setConfig: PropTypes.func.isRequired,
  processorList: PropTypes.object.isRequired,
  processors: PropTypes.array.isRequired,
  setProcessors: PropTypes.func.isRequired,
  channelList: PropTypes.object.isRequired,
  channel: PropTypes.object.isRequired,
  setChannel: PropTypes.func.isRequired,
};

export default ConfigScreen;
