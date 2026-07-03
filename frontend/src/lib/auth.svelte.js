let loggedInUser = $state(null);
let role = $state('');

export const auth = {
  get loggedInUser() {
    return loggedInUser;
  },
  set loggedInUser(val) {
    loggedInUser = val;
  },
  get role() {
    return role;
  },
  set role(val) {
    role = val;
  }
};
