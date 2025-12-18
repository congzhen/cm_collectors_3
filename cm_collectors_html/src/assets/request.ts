import router from '@/router';
import axios, { type AxiosRequestConfig } from 'axios'



//请求前缀
export const requestPrefix = '/api';
const defaultHeaders = {
  'Content-Type': 'application/json'
}

export type IRequest = AxiosRequestConfig
export interface IResponse<T> {
  status: boolean;
  statusCode: number;
  msg: string;
  data: T;
}

axios.interceptors.request.use(
  config => {
    const token = sessionStorage.getItem('token');
    if (token) {
      config.headers.token = token;
    }
    const adminToken = sessionStorage.getItem('adminToken');
    if (adminToken) {
      config.headers.adminToken = adminToken;
    }
    return config;
  }
);

axios.interceptors.response.use(
  response => response,
  error => Promise.reject(error)
);

// eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
function showCustomAlert(str: string, callBack = () => { }, residenceTime = 3000) {
  const alertContainer = document.createElement('div');
  alertContainer.style.display = 'block';
  alertContainer.style.position = 'fixed';
  alertContainer.style.top = '-100%'; // Start off-screen
  alertContainer.style.left = '50%';
  alertContainer.style.transform = 'translate(-50%, -50%)';
  alertContainer.style.width = '10rem';
  alertContainer.style.padding = '1rem';
  alertContainer.style.backgroundColor = 'rgba(68, 72, 74, 0.7)';
  alertContainer.style.borderRadius = '0.5rem';
  alertContainer.style.boxShadow = '0 0 10px rgba(0, 0, 0, 0.3)';
  alertContainer.style.textAlign = 'center';
  alertContainer.style.color = '#FFFFFF';
  alertContainer.style.transition = 'top 0.5s'; // Animation duration
  alertContainer.style.zIndex = '9999';

  const alertMessage = document.createElement('p');
  alertMessage.style.fontSize = '1rem';
  alertMessage.textContent = str;

  alertContainer.appendChild(alertMessage);
  document.body.appendChild(alertContainer);

  // Trigger the animation by setting the top position to the current position
  setTimeout(() => {
    alertContainer.style.top = '10%';
  }, 10); // Delay to allow the element to be rendered before animating
  setTimeout(() => {
    document.body.removeChild(alertContainer);
    callBack();
  }, residenceTime);
}

const _request = <T>(obj: IRequest): Promise<IResponse<T>> => {
  obj.url = requestPrefix + obj.url;
  obj.headers = { ...defaultHeaders, ...(obj.headers || {}) };
  return axios(obj).then(res => {

    if (res.status >= 200 && res.status < 300) {
      return res.data as IResponse<T>;
    }
    //showCustomAlert('Request Error. ' + res.statusText);
    console.error(res.statusText);
    return {
      status: false,
      statusCode: 1,
      msg: res.statusText,
      data: undefined as T
    };
  }).catch(error => {
    if (error.status == 401) {
      router.push('/adminLogin');
    } else {
      //showCustomAlert('Request Error. ' + error.message);
    }
    console.error(error.message);
    return {
      status: false,
      statusCode: 1,
      msg: error.message,
      data: undefined as T
    };
  });
}

export default async <T>(obj: IRequest, contentType: null | 'application/x-www-form-urlencoded' | 'application/json' = null): Promise<IResponse<T>> => {
  try {
    if (contentType != null) {
      obj.headers = {
        'Content-Type': contentType
      }
    }
    return await _request<T>(obj);
  } catch (errorIResponse) {
    return errorIResponse as IResponse<T>;
  }
}

