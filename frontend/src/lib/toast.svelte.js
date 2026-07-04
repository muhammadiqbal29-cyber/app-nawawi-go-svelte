function createToastStore() {
  let messages = $state([]);

  return {
    get messages() {
      return messages;
    },
    show(text, type = 'info', duration = 3000) {
      const id = Date.now() + Math.floor(Math.random() * 1000);
      messages.push({ id, text, type });

      setTimeout(() => {
        messages = messages.filter((m) => m.id !== id);
      }, duration);
    }
  };
}

export const toastStore = createToastStore();
