function createModalStore() {
    let isOpen = $state(false);
    let title = $state('');
    let message = $state('');
    
    let resolvePromise: ((value: boolean) => void) | null = null;

    return {
        get isOpen() { return isOpen; },
        get title() { return title; },
        get message() { return message; },
        
        confirm(modalTitle: string, modalMessage: string): Promise<boolean> {
            title = modalTitle;
            message = modalMessage;
            isOpen = true;
            
            return new Promise((resolve) => {
                resolvePromise = resolve;
            });
        },
        
        handleConfirm() {
            isOpen = false;
            if (resolvePromise) {
                resolvePromise(true);
                resolvePromise = null;
            }
        },
        
        handleCancel() {
            isOpen = false;
            if (resolvePromise) {
                resolvePromise(false);
                resolvePromise = null;
            }
        }
    };
}

export const modalStore = createModalStore();
