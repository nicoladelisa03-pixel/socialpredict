const amountPattern = /^\d+(\.\d{1,2})?$/;

export const reportedValuationDisclaimer = 'This is a reported fundraising valuation from the source below. It is not a prediction-market price or a probability.';

export function majorUnitsToMinor(value) {
  const trimmed = String(value ?? '').trim();
  if (!trimmed) {
    return null;
  }
  if (!amountPattern.test(trimmed)) {
    throw new Error('Enter a positive amount with up to two decimal places, or leave it blank.');
  }
  const [whole, fraction = ''] = trimmed.split('.');
  const minor = (BigInt(whole) * 100n) + BigInt(`${fraction}00`.slice(0, 2));
  if (minor <= 0n || minor > BigInt(Number.MAX_SAFE_INTEGER)) {
    throw new Error('Enter a positive amount with up to two decimal places, or leave it blank.');
  }
  return Number(minor);
}

export function formatMinorAmount(minor, currency) {
  const value = Number(minor);
  if (!Number.isSafeInteger(value)) {
    return '';
  }
  const sign = value < 0 ? '-' : '';
  const absolute = Math.abs(value);
  const whole = Math.trunc(absolute / 100);
  const cents = String(absolute % 100).padStart(2, '0');
  return `${currency} ${sign}${whole.toLocaleString('en-US')}.${cents}`;
}

export function valuationBasisText(basis) {
  if (basis === 'pre-money') {
    return 'pre-money';
  }
  if (basis === 'post-money') {
    return 'post-money';
  }
  return 'basis not stated';
}
