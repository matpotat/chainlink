import { createStore, applyMiddleware, Middleware } from 'redux'
import { composeWithDevTools } from 'redux-devtools-extension'
import thunkMiddleware from 'redux-thunk'
import logger from 'redux-logger'
import reducer from './reducers'

let middleware: Middleware[] = [thunkMiddleware]
// TODO: If the typings for this are too difficult just remove. Can use redux dev tools instead
if (process.env.LOG_REDUX === 'true') {
  middleware = middleware.concat(logger)
}
const composeEnhancers = composeWithDevTools({})

export default () =>
  createStore(reducer, composeEnhancers(applyMiddleware(...middleware)))
