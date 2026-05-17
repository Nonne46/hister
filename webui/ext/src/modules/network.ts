async function fetchFavicon(url) {
  const response = await fetch(url);
  let iconBytes = await response.blob();
  const reader = new FileReader();
  reader.readAsDataURL(iconBytes);
  //let icon = btoa(iconBytes.text());
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onloadend = () => {
      resolve(reader.result);
    };
    reader.onerror = () => resolve('');
    reader.readAsDataURL(iconBytes);
  });
}

async function getServerCookies(): Promise<string> {
  return new Promise((resolve) => {
    chrome.storage.local.get(['histerCookies'], (data) => {
      resolve(data['histerCookies'] || '');
    });
  });
}

async function fetchAPI(
  url: string,
  options: {
    method?: string;
    body?: unknown;
    formData?: Record<string, string>;
    customHeaders?: { name: string; value: string }[];
  } = {},
): Promise<Response> {
  const cookieHeader = await getServerCookies();
  const headers: Record<string, string> = {};

  if (options.body !== undefined) {
    headers['Content-type'] = 'application/json; charset=UTF-8';
  } else if (options.formData !== undefined) {
    headers['Content-type'] = 'application/x-www-form-urlencoded';
  }
  if (cookieHeader) {
    headers['Cookie'] = cookieHeader;
  }
  for (const h of options.customHeaders ?? []) {
    if (h.name) headers[h.name] = h.value || '';
  }

  let fetchBody: BodyInit | undefined;
  if (options.body !== undefined) {
    fetchBody = JSON.stringify(options.body);
  } else if (options.formData !== undefined) {
    fetchBody = new URLSearchParams(options.formData).toString();
  }

  return fetch(url, {
    method: options.method ?? (fetchBody !== undefined ? 'POST' : 'GET'),
    headers,
    body: fetchBody,
    credentials: 'include',
  });
}

async function sendPageData(url, doc, customHeaders = []) {
  try {
    doc['favicon'] = await fetchFavicon(doc.faviconURL);
  } catch (e) {
    doc['favicon'] = '';
  }
  return sendResult(url, doc, customHeaders);
}

async function sendResult(url, res, customHeaders = []) {
  return fetchAPI(url, { body: res, customHeaders });
}

async function sendPDFData(
  url: string,
  doc: Record<string, unknown>,
  pdfBase64: string,
  customHeaders: { name: string; value: string }[] = [],
): Promise<Response> {
  return fetchAPI(url, { body: { document: doc, pdf: pdfBase64 }, customHeaders });
}

export { fetchAPI, sendPageData, sendResult, sendPDFData };
