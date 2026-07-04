function createModalStore() {
  let isOpen = $state(false);
  let title = $state('');
  let message = $state('');
  let resolvePromise = null;

  return {
    get isOpen() {
      return isOpen;
    },
    get title() {
      return title;
    },
    get message() {
      return message;
    },

    handleConfirm() {
      isOpen = false;
      if (resolvePromise) resolvePromise(true);
    },

    handleCancel() {
      isOpen = false;
      if (resolvePromise) resolvePromise(false);
    },

    show(opts) {
      title = opts.title || 'Konfirmasi';
      message = opts.message || '';
      isOpen = true;
      return new Promise((resolve) => {
        resolvePromise = resolve;
      });
    }
  };
}

export const modalStore = createModalStore();
