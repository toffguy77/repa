import 'package:firebase_analytics/firebase_analytics.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class AnalyticsService {
  final FirebaseAnalytics _analytics;

  AnalyticsService(this._analytics);

  // Auth
  Future<void> logSignUp(String method) =>
      _analytics.logSignUp(signUpMethod: method);

  Future<void> logLogin(String method) =>
      _analytics.logLogin(loginMethod: method);

  // Groups
  Future<void> logGroupCreated() =>
      _analytics.logEvent(name: 'group_created');

  Future<void> logGroupJoined() =>
      _analytics.logEvent(name: 'group_joined');

  // Voting
  Future<void> logVotingStarted(String groupId) =>
      _analytics.logEvent(name: 'voting_started', parameters: {
        'group_id': groupId,
      });

  Future<void> logVotingCompleted(String groupId, int questionsAnswered) =>
      _analytics.logEvent(name: 'voting_completed', parameters: {
        'group_id': groupId,
        'questions_answered': questionsAnswered,
      });

  Future<void> logVotingAbandoned(String groupId, int questionsAnswered) =>
      _analytics.logEvent(name: 'voting_abandoned', parameters: {
        'group_id': groupId,
        'questions_answered': questionsAnswered,
      });

  // Reveal
  Future<void> logRevealOpened(String groupId) =>
      _analytics.logEvent(name: 'reveal_opened', parameters: {
        'group_id': groupId,
      });

  Future<void> logCardShared(String method) =>
      _analytics.logEvent(name: 'card_shared', parameters: {
        'method': method,
      });

  Future<void> logDetectorPurchased() =>
      _analytics.logEvent(name: 'detector_purchased');

  Future<void> logHiddenAttributesOpened() =>
      _analytics.logEvent(name: 'hidden_attributes_opened');

  // Monetization
  Future<void> logPurchaseInitiated(String packageId, int priceKopecks) =>
      _analytics.logEvent(name: 'purchase_initiated', parameters: {
        'package_id': packageId,
        'price_kopecks': priceKopecks,
      });

  Future<void> logPurchaseCompleted(String packageId, int crystals) =>
      _analytics.logEvent(name: 'purchase_completed', parameters: {
        'package_id': packageId,
        'crystals': crystals,
      });

  // Achievements
  Future<void> logAchievementUnlocked(String achievementType) =>
      _analytics.logEvent(name: 'achievement_unlocked', parameters: {
        'achievement_type': achievementType,
      });

  // Retention
  Future<void> logReactionSent() =>
      _analytics.logEvent(name: 'reaction_sent');

  Future<void> logQuestionVoted() =>
      _analytics.logEvent(name: 'question_voted');
}

final analyticsProvider = Provider<AnalyticsService>((ref) {
  return AnalyticsService(FirebaseAnalytics.instance);
});
