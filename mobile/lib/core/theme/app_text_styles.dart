import 'package:flutter/material.dart';
import 'app_colors.dart';

class AppTextStyles {
  static const headline1 = TextStyle(
    fontSize: 28,
    fontWeight: FontWeight.bold,
    color: AppColors.textPrimary,
  );

  static const headline2 = TextStyle(
    fontSize: 22,
    fontWeight: FontWeight.bold,
    color: AppColors.textPrimary,
  );

  static const body = TextStyle(
    fontSize: 16,
    color: AppColors.textPrimary,
  );

  static const bodySecondary = TextStyle(
    fontSize: 16,
    color: AppColors.textSecondary,
  );

  static const caption = TextStyle(
    fontSize: 14,
    color: AppColors.textSecondary,
  );

  static const button = TextStyle(
    fontSize: 16,
    fontWeight: FontWeight.w600,
    color: Colors.white,
  );
}
