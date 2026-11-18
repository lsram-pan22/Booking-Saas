export interface ResettableOwner { id: string; full_name: string; email: string; username: string }
export interface PasswordForm { password: string; confirm_password: string }
export interface ResetPasswordDrawerProps {
  name: string;
  open: boolean;
  onClose: () => void;
  onReset: (password: string, currentPassword: string) => Promise<void>;
}
export interface ChangePasswordForm extends PasswordForm { current_password: string }
export interface ResetPasswordForm extends PasswordForm { current_password: string }
