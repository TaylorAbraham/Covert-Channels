import React from 'react';
import renderer from 'react-test-renderer';
import App from '../App';

describe('Initialization', () => {
  it('requests an initial config and parses the response correctly', () => {
    const app = renderer.create(<App />);
    expect(app.toJSON()).toMatchSnapshot();
  });
});
