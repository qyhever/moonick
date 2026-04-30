type ToastProps = {
  message: string;
};

export default function Toast({ message }: ToastProps) {
  return (
    <div aria-live="polite" className="toast-banner" role="status">
      {message}
    </div>
  );
}
