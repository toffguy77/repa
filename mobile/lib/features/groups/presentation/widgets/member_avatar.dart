import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../../core/theme/app_colors.dart';

class MemberAvatar extends StatelessWidget {
  final String? avatarEmoji;
  final String? avatarUrl;
  final double size;
  final int votingStreak;

  const MemberAvatar({
    super.key,
    this.avatarEmoji,
    this.avatarUrl,
    this.size = 44,
    this.votingStreak = 0,
  });

  @override
  Widget build(BuildContext context) {
    return Stack(
      clipBehavior: Clip.none,
      children: [
        Container(
          width: size,
          height: size,
          decoration: BoxDecoration(
            color: AppColors.primaryLight,
            shape: BoxShape.circle,
            image: avatarUrl != null
                ? DecorationImage(
                    image: CachedNetworkImageProvider(avatarUrl!),
                    fit: BoxFit.cover,
                  )
                : null,
          ),
          child: avatarUrl == null
              ? Center(
                  child: Text(
                    avatarEmoji ?? '?',
                    style: TextStyle(fontSize: size * 0.5),
                  ),
                )
              : null,
        ),
        if (votingStreak >= 3)
          Positioned(
            right: -4,
            bottom: -4,
            child: Container(
              padding: const EdgeInsets.all(2),
              decoration: const BoxDecoration(
                color: Colors.white,
                shape: BoxShape.circle,
              ),
              child: Text(
                '\u{1F525}',
                style: TextStyle(fontSize: size * 0.3),
              ),
            ),
          ),
      ],
    );
  }
}
