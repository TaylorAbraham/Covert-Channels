// The warning is disabled because this is a dev dependency,
// but ESLint sees it as a regular dependency
/* eslint-disable import/no-extraneous-dependencies */
import Enzyme from 'enzyme';
import Adapter from 'enzyme-adapter-react-16';

Enzyme.configure({ adapter: new Adapter() });
