import { isAuthenticated } from '../services/auth';
import { router } from 'tinro';

export const routerState = $state({
    navParams: null as any
});

export function navigate(view: string, payload?: any) {
    // Auth guard
    if (view !== 'login' && !isAuthenticated()) {
        router.goto('/login');
        return;
    }

    if (payload) {
        routerState.navParams = payload;
    } else {
        routerState.navParams = null;
    }
    
    // tinro routing
    const target = view === 'splash' ? '/' : '/' + view;
    router.goto(target);
}
