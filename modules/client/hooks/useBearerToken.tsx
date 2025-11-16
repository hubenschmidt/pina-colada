import { useContext } from 'react';
import { UserContext } from '../context/userContext';

export const useBearerToken = () => {
    const {
        userState: { bearerToken },
    } = useContext(UserContext);
    return bearerToken;
};
