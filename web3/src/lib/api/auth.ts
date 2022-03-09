import state from '$lib/state'

let auth = { ...state.initialAuthState };
state.auth.subscribe(authState => {
  auth = authState
});

const getAuth = async ({ authState }) => {
  if (!authState) {
    const token = auth.token
    if (token && token != '') {
      return { token };
    }
    return null;
  }

  return null;
};

const willAuthError = ({ authState }) => {
  if (!authState || !authState.token) return true;
  return false;
}

const addAuthToOperation = ({ authState, operation }) => {
  // the token isn't in the auth state, return the operation without changes
  if (!authState || !authState.token) {
    return operation;
  }

  // fetchOptions can be a function (See Client API) but you can simplify this based on usage
  const fetchOptions =
    typeof operation.context.fetchOptions === 'function'
      ? operation.context.fetchOptions()
      : operation.context.fetchOptions || {};

  return {
    ...operation,
    context: {
      ...operation.context,
      fetchOptions: {
        ...fetchOptions,
        headers: {
          ...fetchOptions.headers,
          "Authorization": `token ${authState.token}`,
        },
      },
    },
  };
}

export default {
  getAuth,
  willAuthError,
  addAuthToOperation,
}