import React from 'react';
import PropTypes from 'prop-types';
import Button from 'react-bootstrap/Button';
import Dropdown from 'react-bootstrap/Dropdown';

import Checkbox from '../ui-components/Checkbox';
import IPInput from '../ui-components/IPInput';
import NumberInput from '../ui-components/NumberInput';
import Select from '../ui-components/Select';
import TextArea from '../ui-components/TextArea';

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
    channelIsOpen,
  } = props;

  return (
    <div className="m-2">
      <h2 className="m-1">Configuration</h2>
      <h3 className="m-1">Processors</h3>
      <Button
        variant="success"
        className="m-1 w-100"
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
                className="w-100"
                variant="outline-primary"
              >
                {processor.Type || 'Select a Processor'}
              </Dropdown.Toggle>
              <Dropdown.Menu className="w-100">
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
              let propsForComponent = {
                key,
                label: opt.Display.Name,
                value: opt.Value,
                tooltip: opt.Display.Description,
                parentOnChange: e => setProcessors([
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
                    <IPInput {...propsForComponent} />
                  );
                case 'i8':
                case 'u16':
                case 'u64':
                  if (opt.Range) {
                    propsForComponent = {
                      ...propsForComponent,
                      min: opt.Range[0],
                      max: opt.Range[1],
                    };
                  }
                  return (
                    <NumberInput
                      {...propsForComponent}
                      parentOnChange={e => setProcessors([
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
                case 'exactu64':
                  return (<div>EXACTU64</div>);
                case 'bool':
                  return (
                    <Checkbox
                      {...propsForComponent}
                      parentOnChange={e => setProcessors([
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
                      {...propsForComponent}
                      items={opt.Range}
                    />
                  );
                case 'hexkey':
                  return (<div>hexkey</div>);
                case 'key':
                  return (<TextArea {...propsForComponent} />);
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
          className="cc-config__chan-select w-100"
          variant="outline-primary"
        >
          {channel.value || 'Select a Channel'}
        </Dropdown.Toggle>
        <Dropdown.Menu className="w-100">
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
        let propsForComponent = {
          key,
          label: opt.Display.Name,
          value: opt.Value,
          tooltip: opt.Display.Description,
          parentOnChange: e => setConfig({
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
              <IPInput {...propsForComponent} />
            );
          case 'i8':
          case 'u16':
          case 'u64':
            if (opt.Range) {
              propsForComponent = {
                ...propsForComponent,
                min: opt.Range[0],
                max: opt.Range[1],
              };
            }
            return (
              <NumberInput
                {...propsForComponent}
                parentOnChange={e => setConfig({
                  ...config,
                  [key]: {
                    ...config[key],
                    Value: parseInt(e.target.value) || 0,
                  },
                })}
              />
            );
          case 'exactu64':
            return (<div>EXACTU64</div>);
          case 'bool':
            return (
              <Checkbox
                {...propsForComponent}
                parentOnChange={e => setConfig({
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
                {...propsForComponent}
                items={opt.Range}
              />
            );
          case 'hexkey':
            return (<div>hexkey</div>);
          case 'key':
            return (<TextArea {...propsForComponent} />);
          default:
            return (<div key={key}>UNIMPLEMENTED</div>);
        }
      })}
      {channelIsOpen ? (
        <Button variant="danger" onClick={closeChannel} className="m-1 w-100">Close Covert Channel</Button>
      ) : (
        <Button
          variant="success"
          onClick={openChannel}
          className="cc-config__submit m-1 w-100"
          hidden={Object.entries(channel).length === 0 && channel.constructor === Object}
        >
          Open Covert Channel
        </Button>
      )}
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
  channelIsOpen: PropTypes.bool.isRequired,
};

export default ConfigScreen;
