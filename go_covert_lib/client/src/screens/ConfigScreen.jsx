import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import FileSaver from 'file-saver';
import Button from 'react-bootstrap/Button';
import Dropdown from 'react-bootstrap/Dropdown';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash } from '@fortawesome/free-solid-svg-icons';

import Checkbox from '../ui-components/Checkbox';
import HexKey from '../ui-components/HexKey';
import IPInput from '../ui-components/IPInput';
import NumberInput from '../ui-components/NumberInput';
import Select from '../ui-components/Select';
import StringInput from '../ui-components/StringInput';
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
    addSystemMessage,
  } = props;

  const [channelIsSelected, setChannelIsSelected] = useState(false);

  const deleteProcessor = (targetIndex) => {
    setProcessors(processors.filter((proc, i) => {
      return i !== targetIndex;
    }));
  };

  const saveConfig = () => {
    const serializedConfig = JSON.stringify({
      config,
      processors,
    });
    const blob = new Blob([serializedConfig], { type: 'text/plain;charset=utf-8' });
    FileSaver.saveAs(blob, 'covert-config.txt');
    addSystemMessage('Saved current configuration');
  };

  const handleFileChange = (e) => {
    if (e && e.target && e.target.files && e.target.files.length === 1) {
      const file = e.target.files[0];
      addSystemMessage(`Loading configuration from ${file.name}...`);
      const reader = new FileReader();
      reader.onload = () => {
        const text = reader.result;
        let parsedConfig;
        try {
          parsedConfig = JSON.parse(text);
          setConfig(parsedConfig.config);
          setProcessors(parsedConfig.processors);
        } catch (err) {
          addSystemMessage(`Could not parse file ${file.name}`);
          return;
        }
        addSystemMessage(`Succesfully loaded configuration from ${file.name}`);
      };
      reader.readAsText(file);
      // Hack to clear the uploaded file. This is used because otherwise uploading
      // the same file will NOT re-load the config. We do want a reload here.
      document.getElementById('invisible-file-input').type = 'text';
      document.getElementById('invisible-file-input').type = 'file';
    }
  };

  useEffect(() => {
    setChannelIsSelected(Object.entries(channel).length > 0 && channel.constructor === Object);
  }, [channel]);

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
          <div className="cc-processor" key={i.toString()}>
            <div className="d-flex">
              <Dropdown className="w-100 m-1">
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
              <Button
                onClick={() => deleteProcessor(i)}
                variant="danger"
                className="cc-processor cc-processor__delete"
              >
                <FontAwesomeIcon icon={faTrash} />
              </Button>
            </div>
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
                  return (<StringInput {...propsForComponent} />);
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
                  return (<HexKey {...propsForComponent} acceptedLengths={opt.Range} />);
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
            return (<StringInput {...propsForComponent} />);
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
            return (<HexKey {...propsForComponent} />);
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
          hidden={!channelIsSelected}
        >
          Open Covert Channel
        </Button>
      )}
      <div className="m-1 w-100 d-flex">
        <Button
          className="mr-1 w-100"
          onClick={saveConfig}
        >
          Save Config
        </Button>
        <Button
          className="w-100"
          onClick={() => document.getElementById('invisible-file-input').click()}
          disabled={channelIsOpen}
        >
          Load Config
        </Button>
        <input onChange={handleFileChange} id="invisible-file-input" type="file" style={{ display: 'none' }} />
      </div>
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
  addSystemMessage: PropTypes.func.isRequired,
};

export default ConfigScreen;
