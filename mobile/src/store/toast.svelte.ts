export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface ToastMessage {
    id: number;
    text: string;
    type: ToastType;
}

function createToastStore() {
    let messages = $state<ToastMessage[]>([]);

    return {
        get messages() { return messages; },
        show(text: string, type: ToastType = 'info', duration: number = 3000) {
            const id = Date.now() + Math.floor(Math.random() * 1000);
            messages.push({ id, text, type });
            
            setTimeout(() => {
                messages = messages.filter(m => m.id !== id);
            }, duration);
        }
    };
}

export const toastStore = createToastStore();
