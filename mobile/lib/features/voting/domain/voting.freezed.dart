// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'voting.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

VotingQuestion _$VotingQuestionFromJson(Map<String, dynamic> json) {
  return _VotingQuestion.fromJson(json);
}

/// @nodoc
mixin _$VotingQuestion {
  @JsonKey(name: 'question_id')
  String get questionId => throw _privateConstructorUsedError;
  String get text => throw _privateConstructorUsedError;
  String get category => throw _privateConstructorUsedError;
  bool get answered => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VotingQuestionCopyWith<VotingQuestion> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VotingQuestionCopyWith<$Res> {
  factory $VotingQuestionCopyWith(
          VotingQuestion value, $Res Function(VotingQuestion) then) =
      _$VotingQuestionCopyWithImpl<$Res, VotingQuestion>;
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      String text,
      String category,
      bool answered});
}

/// @nodoc
class _$VotingQuestionCopyWithImpl<$Res, $Val extends VotingQuestion>
    implements $VotingQuestionCopyWith<$Res> {
  _$VotingQuestionCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? text = null,
    Object? category = null,
    Object? answered = null,
  }) {
    return _then(_value.copyWith(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      text: null == text
          ? _value.text
          : text // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      answered: null == answered
          ? _value.answered
          : answered // ignore: cast_nullable_to_non_nullable
              as bool,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$VotingQuestionImplCopyWith<$Res>
    implements $VotingQuestionCopyWith<$Res> {
  factory _$$VotingQuestionImplCopyWith(_$VotingQuestionImpl value,
          $Res Function(_$VotingQuestionImpl) then) =
      __$$VotingQuestionImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      String text,
      String category,
      bool answered});
}

/// @nodoc
class __$$VotingQuestionImplCopyWithImpl<$Res>
    extends _$VotingQuestionCopyWithImpl<$Res, _$VotingQuestionImpl>
    implements _$$VotingQuestionImplCopyWith<$Res> {
  __$$VotingQuestionImplCopyWithImpl(
      _$VotingQuestionImpl _value, $Res Function(_$VotingQuestionImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? text = null,
    Object? category = null,
    Object? answered = null,
  }) {
    return _then(_$VotingQuestionImpl(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      text: null == text
          ? _value.text
          : text // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      answered: null == answered
          ? _value.answered
          : answered // ignore: cast_nullable_to_non_nullable
              as bool,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VotingQuestionImpl implements _VotingQuestion {
  const _$VotingQuestionImpl(
      {@JsonKey(name: 'question_id') required this.questionId,
      required this.text,
      required this.category,
      required this.answered});

  factory _$VotingQuestionImpl.fromJson(Map<String, dynamic> json) =>
      _$$VotingQuestionImplFromJson(json);

  @override
  @JsonKey(name: 'question_id')
  final String questionId;
  @override
  final String text;
  @override
  final String category;
  @override
  final bool answered;

  @override
  String toString() {
    return 'VotingQuestion(questionId: $questionId, text: $text, category: $category, answered: $answered)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VotingQuestionImpl &&
            (identical(other.questionId, questionId) ||
                other.questionId == questionId) &&
            (identical(other.text, text) || other.text == text) &&
            (identical(other.category, category) ||
                other.category == category) &&
            (identical(other.answered, answered) ||
                other.answered == answered));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, questionId, text, category, answered);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VotingQuestionImplCopyWith<_$VotingQuestionImpl> get copyWith =>
      __$$VotingQuestionImplCopyWithImpl<_$VotingQuestionImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VotingQuestionImplToJson(
      this,
    );
  }
}

abstract class _VotingQuestion implements VotingQuestion {
  const factory _VotingQuestion(
      {@JsonKey(name: 'question_id') required final String questionId,
      required final String text,
      required final String category,
      required final bool answered}) = _$VotingQuestionImpl;

  factory _VotingQuestion.fromJson(Map<String, dynamic> json) =
      _$VotingQuestionImpl.fromJson;

  @override
  @JsonKey(name: 'question_id')
  String get questionId;
  @override
  String get text;
  @override
  String get category;
  @override
  bool get answered;
  @override
  @JsonKey(ignore: true)
  _$$VotingQuestionImplCopyWith<_$VotingQuestionImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

VotingTarget _$VotingTargetFromJson(Map<String, dynamic> json) {
  return _VotingTarget.fromJson(json);
}

/// @nodoc
mixin _$VotingTarget {
  @JsonKey(name: 'user_id')
  String get userId => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VotingTargetCopyWith<VotingTarget> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VotingTargetCopyWith<$Res> {
  factory $VotingTargetCopyWith(
          VotingTarget value, $Res Function(VotingTarget) then) =
      _$VotingTargetCopyWithImpl<$Res, VotingTarget>;
  @useResult
  $Res call(
      {@JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl});
}

/// @nodoc
class _$VotingTargetCopyWithImpl<$Res, $Val extends VotingTarget>
    implements $VotingTargetCopyWith<$Res> {
  _$VotingTargetCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
  }) {
    return _then(_value.copyWith(
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$VotingTargetImplCopyWith<$Res>
    implements $VotingTargetCopyWith<$Res> {
  factory _$$VotingTargetImplCopyWith(
          _$VotingTargetImpl value, $Res Function(_$VotingTargetImpl) then) =
      __$$VotingTargetImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl});
}

/// @nodoc
class __$$VotingTargetImplCopyWithImpl<$Res>
    extends _$VotingTargetCopyWithImpl<$Res, _$VotingTargetImpl>
    implements _$$VotingTargetImplCopyWith<$Res> {
  __$$VotingTargetImplCopyWithImpl(
      _$VotingTargetImpl _value, $Res Function(_$VotingTargetImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
  }) {
    return _then(_$VotingTargetImpl(
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VotingTargetImpl implements _VotingTarget {
  const _$VotingTargetImpl(
      {@JsonKey(name: 'user_id') required this.userId,
      required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      @JsonKey(name: 'avatar_url') this.avatarUrl});

  factory _$VotingTargetImpl.fromJson(Map<String, dynamic> json) =>
      _$$VotingTargetImplFromJson(json);

  @override
  @JsonKey(name: 'user_id')
  final String userId;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  final String? avatarUrl;

  @override
  String toString() {
    return 'VotingTarget(userId: $userId, username: $username, avatarEmoji: $avatarEmoji, avatarUrl: $avatarUrl)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VotingTargetImpl &&
            (identical(other.userId, userId) || other.userId == userId) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.avatarUrl, avatarUrl) ||
                other.avatarUrl == avatarUrl));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, userId, username, avatarEmoji, avatarUrl);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VotingTargetImplCopyWith<_$VotingTargetImpl> get copyWith =>
      __$$VotingTargetImplCopyWithImpl<_$VotingTargetImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VotingTargetImplToJson(
      this,
    );
  }
}

abstract class _VotingTarget implements VotingTarget {
  const factory _VotingTarget(
          {@JsonKey(name: 'user_id') required final String userId,
          required final String username,
          @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
          @JsonKey(name: 'avatar_url') final String? avatarUrl}) =
      _$VotingTargetImpl;

  factory _VotingTarget.fromJson(Map<String, dynamic> json) =
      _$VotingTargetImpl.fromJson;

  @override
  @JsonKey(name: 'user_id')
  String get userId;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl;
  @override
  @JsonKey(ignore: true)
  _$$VotingTargetImplCopyWith<_$VotingTargetImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

VotingProgress _$VotingProgressFromJson(Map<String, dynamic> json) {
  return _VotingProgress.fromJson(json);
}

/// @nodoc
mixin _$VotingProgress {
  int get answered => throw _privateConstructorUsedError;
  int get total => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VotingProgressCopyWith<VotingProgress> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VotingProgressCopyWith<$Res> {
  factory $VotingProgressCopyWith(
          VotingProgress value, $Res Function(VotingProgress) then) =
      _$VotingProgressCopyWithImpl<$Res, VotingProgress>;
  @useResult
  $Res call({int answered, int total});
}

/// @nodoc
class _$VotingProgressCopyWithImpl<$Res, $Val extends VotingProgress>
    implements $VotingProgressCopyWith<$Res> {
  _$VotingProgressCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? answered = null,
    Object? total = null,
  }) {
    return _then(_value.copyWith(
      answered: null == answered
          ? _value.answered
          : answered // ignore: cast_nullable_to_non_nullable
              as int,
      total: null == total
          ? _value.total
          : total // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$VotingProgressImplCopyWith<$Res>
    implements $VotingProgressCopyWith<$Res> {
  factory _$$VotingProgressImplCopyWith(_$VotingProgressImpl value,
          $Res Function(_$VotingProgressImpl) then) =
      __$$VotingProgressImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({int answered, int total});
}

/// @nodoc
class __$$VotingProgressImplCopyWithImpl<$Res>
    extends _$VotingProgressCopyWithImpl<$Res, _$VotingProgressImpl>
    implements _$$VotingProgressImplCopyWith<$Res> {
  __$$VotingProgressImplCopyWithImpl(
      _$VotingProgressImpl _value, $Res Function(_$VotingProgressImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? answered = null,
    Object? total = null,
  }) {
    return _then(_$VotingProgressImpl(
      answered: null == answered
          ? _value.answered
          : answered // ignore: cast_nullable_to_non_nullable
              as int,
      total: null == total
          ? _value.total
          : total // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VotingProgressImpl implements _VotingProgress {
  const _$VotingProgressImpl({required this.answered, required this.total});

  factory _$VotingProgressImpl.fromJson(Map<String, dynamic> json) =>
      _$$VotingProgressImplFromJson(json);

  @override
  final int answered;
  @override
  final int total;

  @override
  String toString() {
    return 'VotingProgress(answered: $answered, total: $total)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VotingProgressImpl &&
            (identical(other.answered, answered) ||
                other.answered == answered) &&
            (identical(other.total, total) || other.total == total));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, answered, total);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VotingProgressImplCopyWith<_$VotingProgressImpl> get copyWith =>
      __$$VotingProgressImplCopyWithImpl<_$VotingProgressImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VotingProgressImplToJson(
      this,
    );
  }
}

abstract class _VotingProgress implements VotingProgress {
  const factory _VotingProgress(
      {required final int answered,
      required final int total}) = _$VotingProgressImpl;

  factory _VotingProgress.fromJson(Map<String, dynamic> json) =
      _$VotingProgressImpl.fromJson;

  @override
  int get answered;
  @override
  int get total;
  @override
  @JsonKey(ignore: true)
  _$$VotingProgressImplCopyWith<_$VotingProgressImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

VotingSession _$VotingSessionFromJson(Map<String, dynamic> json) {
  return _VotingSession.fromJson(json);
}

/// @nodoc
mixin _$VotingSession {
  @JsonKey(name: 'season_id')
  String get seasonId => throw _privateConstructorUsedError;
  List<VotingQuestion> get questions => throw _privateConstructorUsedError;
  List<VotingTarget> get targets => throw _privateConstructorUsedError;
  VotingProgress get progress => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VotingSessionCopyWith<VotingSession> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VotingSessionCopyWith<$Res> {
  factory $VotingSessionCopyWith(
          VotingSession value, $Res Function(VotingSession) then) =
      _$VotingSessionCopyWithImpl<$Res, VotingSession>;
  @useResult
  $Res call(
      {@JsonKey(name: 'season_id') String seasonId,
      List<VotingQuestion> questions,
      List<VotingTarget> targets,
      VotingProgress progress});

  $VotingProgressCopyWith<$Res> get progress;
}

/// @nodoc
class _$VotingSessionCopyWithImpl<$Res, $Val extends VotingSession>
    implements $VotingSessionCopyWith<$Res> {
  _$VotingSessionCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonId = null,
    Object? questions = null,
    Object? targets = null,
    Object? progress = null,
  }) {
    return _then(_value.copyWith(
      seasonId: null == seasonId
          ? _value.seasonId
          : seasonId // ignore: cast_nullable_to_non_nullable
              as String,
      questions: null == questions
          ? _value.questions
          : questions // ignore: cast_nullable_to_non_nullable
              as List<VotingQuestion>,
      targets: null == targets
          ? _value.targets
          : targets // ignore: cast_nullable_to_non_nullable
              as List<VotingTarget>,
      progress: null == progress
          ? _value.progress
          : progress // ignore: cast_nullable_to_non_nullable
              as VotingProgress,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $VotingProgressCopyWith<$Res> get progress {
    return $VotingProgressCopyWith<$Res>(_value.progress, (value) {
      return _then(_value.copyWith(progress: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$VotingSessionImplCopyWith<$Res>
    implements $VotingSessionCopyWith<$Res> {
  factory _$$VotingSessionImplCopyWith(
          _$VotingSessionImpl value, $Res Function(_$VotingSessionImpl) then) =
      __$$VotingSessionImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'season_id') String seasonId,
      List<VotingQuestion> questions,
      List<VotingTarget> targets,
      VotingProgress progress});

  @override
  $VotingProgressCopyWith<$Res> get progress;
}

/// @nodoc
class __$$VotingSessionImplCopyWithImpl<$Res>
    extends _$VotingSessionCopyWithImpl<$Res, _$VotingSessionImpl>
    implements _$$VotingSessionImplCopyWith<$Res> {
  __$$VotingSessionImplCopyWithImpl(
      _$VotingSessionImpl _value, $Res Function(_$VotingSessionImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonId = null,
    Object? questions = null,
    Object? targets = null,
    Object? progress = null,
  }) {
    return _then(_$VotingSessionImpl(
      seasonId: null == seasonId
          ? _value.seasonId
          : seasonId // ignore: cast_nullable_to_non_nullable
              as String,
      questions: null == questions
          ? _value._questions
          : questions // ignore: cast_nullable_to_non_nullable
              as List<VotingQuestion>,
      targets: null == targets
          ? _value._targets
          : targets // ignore: cast_nullable_to_non_nullable
              as List<VotingTarget>,
      progress: null == progress
          ? _value.progress
          : progress // ignore: cast_nullable_to_non_nullable
              as VotingProgress,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VotingSessionImpl implements _VotingSession {
  const _$VotingSessionImpl(
      {@JsonKey(name: 'season_id') required this.seasonId,
      required final List<VotingQuestion> questions,
      required final List<VotingTarget> targets,
      required this.progress})
      : _questions = questions,
        _targets = targets;

  factory _$VotingSessionImpl.fromJson(Map<String, dynamic> json) =>
      _$$VotingSessionImplFromJson(json);

  @override
  @JsonKey(name: 'season_id')
  final String seasonId;
  final List<VotingQuestion> _questions;
  @override
  List<VotingQuestion> get questions {
    if (_questions is EqualUnmodifiableListView) return _questions;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_questions);
  }

  final List<VotingTarget> _targets;
  @override
  List<VotingTarget> get targets {
    if (_targets is EqualUnmodifiableListView) return _targets;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_targets);
  }

  @override
  final VotingProgress progress;

  @override
  String toString() {
    return 'VotingSession(seasonId: $seasonId, questions: $questions, targets: $targets, progress: $progress)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VotingSessionImpl &&
            (identical(other.seasonId, seasonId) ||
                other.seasonId == seasonId) &&
            const DeepCollectionEquality()
                .equals(other._questions, _questions) &&
            const DeepCollectionEquality().equals(other._targets, _targets) &&
            (identical(other.progress, progress) ||
                other.progress == progress));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      seasonId,
      const DeepCollectionEquality().hash(_questions),
      const DeepCollectionEquality().hash(_targets),
      progress);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VotingSessionImplCopyWith<_$VotingSessionImpl> get copyWith =>
      __$$VotingSessionImplCopyWithImpl<_$VotingSessionImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VotingSessionImplToJson(
      this,
    );
  }
}

abstract class _VotingSession implements VotingSession {
  const factory _VotingSession(
      {@JsonKey(name: 'season_id') required final String seasonId,
      required final List<VotingQuestion> questions,
      required final List<VotingTarget> targets,
      required final VotingProgress progress}) = _$VotingSessionImpl;

  factory _VotingSession.fromJson(Map<String, dynamic> json) =
      _$VotingSessionImpl.fromJson;

  @override
  @JsonKey(name: 'season_id')
  String get seasonId;
  @override
  List<VotingQuestion> get questions;
  @override
  List<VotingTarget> get targets;
  @override
  VotingProgress get progress;
  @override
  @JsonKey(ignore: true)
  _$$VotingSessionImplCopyWith<_$VotingSessionImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

VoteResultData _$VoteResultDataFromJson(Map<String, dynamic> json) {
  return _VoteResultData.fromJson(json);
}

/// @nodoc
mixin _$VoteResultData {
  VoteInfo get vote => throw _privateConstructorUsedError;
  VotingProgress get progress => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VoteResultDataCopyWith<VoteResultData> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VoteResultDataCopyWith<$Res> {
  factory $VoteResultDataCopyWith(
          VoteResultData value, $Res Function(VoteResultData) then) =
      _$VoteResultDataCopyWithImpl<$Res, VoteResultData>;
  @useResult
  $Res call({VoteInfo vote, VotingProgress progress});

  $VoteInfoCopyWith<$Res> get vote;
  $VotingProgressCopyWith<$Res> get progress;
}

/// @nodoc
class _$VoteResultDataCopyWithImpl<$Res, $Val extends VoteResultData>
    implements $VoteResultDataCopyWith<$Res> {
  _$VoteResultDataCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? vote = null,
    Object? progress = null,
  }) {
    return _then(_value.copyWith(
      vote: null == vote
          ? _value.vote
          : vote // ignore: cast_nullable_to_non_nullable
              as VoteInfo,
      progress: null == progress
          ? _value.progress
          : progress // ignore: cast_nullable_to_non_nullable
              as VotingProgress,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $VoteInfoCopyWith<$Res> get vote {
    return $VoteInfoCopyWith<$Res>(_value.vote, (value) {
      return _then(_value.copyWith(vote: value) as $Val);
    });
  }

  @override
  @pragma('vm:prefer-inline')
  $VotingProgressCopyWith<$Res> get progress {
    return $VotingProgressCopyWith<$Res>(_value.progress, (value) {
      return _then(_value.copyWith(progress: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$VoteResultDataImplCopyWith<$Res>
    implements $VoteResultDataCopyWith<$Res> {
  factory _$$VoteResultDataImplCopyWith(_$VoteResultDataImpl value,
          $Res Function(_$VoteResultDataImpl) then) =
      __$$VoteResultDataImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({VoteInfo vote, VotingProgress progress});

  @override
  $VoteInfoCopyWith<$Res> get vote;
  @override
  $VotingProgressCopyWith<$Res> get progress;
}

/// @nodoc
class __$$VoteResultDataImplCopyWithImpl<$Res>
    extends _$VoteResultDataCopyWithImpl<$Res, _$VoteResultDataImpl>
    implements _$$VoteResultDataImplCopyWith<$Res> {
  __$$VoteResultDataImplCopyWithImpl(
      _$VoteResultDataImpl _value, $Res Function(_$VoteResultDataImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? vote = null,
    Object? progress = null,
  }) {
    return _then(_$VoteResultDataImpl(
      vote: null == vote
          ? _value.vote
          : vote // ignore: cast_nullable_to_non_nullable
              as VoteInfo,
      progress: null == progress
          ? _value.progress
          : progress // ignore: cast_nullable_to_non_nullable
              as VotingProgress,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VoteResultDataImpl implements _VoteResultData {
  const _$VoteResultDataImpl({required this.vote, required this.progress});

  factory _$VoteResultDataImpl.fromJson(Map<String, dynamic> json) =>
      _$$VoteResultDataImplFromJson(json);

  @override
  final VoteInfo vote;
  @override
  final VotingProgress progress;

  @override
  String toString() {
    return 'VoteResultData(vote: $vote, progress: $progress)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VoteResultDataImpl &&
            (identical(other.vote, vote) || other.vote == vote) &&
            (identical(other.progress, progress) ||
                other.progress == progress));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, vote, progress);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VoteResultDataImplCopyWith<_$VoteResultDataImpl> get copyWith =>
      __$$VoteResultDataImplCopyWithImpl<_$VoteResultDataImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VoteResultDataImplToJson(
      this,
    );
  }
}

abstract class _VoteResultData implements VoteResultData {
  const factory _VoteResultData(
      {required final VoteInfo vote,
      required final VotingProgress progress}) = _$VoteResultDataImpl;

  factory _VoteResultData.fromJson(Map<String, dynamic> json) =
      _$VoteResultDataImpl.fromJson;

  @override
  VoteInfo get vote;
  @override
  VotingProgress get progress;
  @override
  @JsonKey(ignore: true)
  _$$VoteResultDataImplCopyWith<_$VoteResultDataImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

VoteInfo _$VoteInfoFromJson(Map<String, dynamic> json) {
  return _VoteInfo.fromJson(json);
}

/// @nodoc
mixin _$VoteInfo {
  @JsonKey(name: 'question_id')
  String get questionId => throw _privateConstructorUsedError;
  @JsonKey(name: 'target_id')
  String get targetId => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VoteInfoCopyWith<VoteInfo> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VoteInfoCopyWith<$Res> {
  factory $VoteInfoCopyWith(VoteInfo value, $Res Function(VoteInfo) then) =
      _$VoteInfoCopyWithImpl<$Res, VoteInfo>;
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'target_id') String targetId});
}

/// @nodoc
class _$VoteInfoCopyWithImpl<$Res, $Val extends VoteInfo>
    implements $VoteInfoCopyWith<$Res> {
  _$VoteInfoCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? targetId = null,
  }) {
    return _then(_value.copyWith(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      targetId: null == targetId
          ? _value.targetId
          : targetId // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$VoteInfoImplCopyWith<$Res>
    implements $VoteInfoCopyWith<$Res> {
  factory _$$VoteInfoImplCopyWith(
          _$VoteInfoImpl value, $Res Function(_$VoteInfoImpl) then) =
      __$$VoteInfoImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'target_id') String targetId});
}

/// @nodoc
class __$$VoteInfoImplCopyWithImpl<$Res>
    extends _$VoteInfoCopyWithImpl<$Res, _$VoteInfoImpl>
    implements _$$VoteInfoImplCopyWith<$Res> {
  __$$VoteInfoImplCopyWithImpl(
      _$VoteInfoImpl _value, $Res Function(_$VoteInfoImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? targetId = null,
  }) {
    return _then(_$VoteInfoImpl(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      targetId: null == targetId
          ? _value.targetId
          : targetId // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VoteInfoImpl implements _VoteInfo {
  const _$VoteInfoImpl(
      {@JsonKey(name: 'question_id') required this.questionId,
      @JsonKey(name: 'target_id') required this.targetId});

  factory _$VoteInfoImpl.fromJson(Map<String, dynamic> json) =>
      _$$VoteInfoImplFromJson(json);

  @override
  @JsonKey(name: 'question_id')
  final String questionId;
  @override
  @JsonKey(name: 'target_id')
  final String targetId;

  @override
  String toString() {
    return 'VoteInfo(questionId: $questionId, targetId: $targetId)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VoteInfoImpl &&
            (identical(other.questionId, questionId) ||
                other.questionId == questionId) &&
            (identical(other.targetId, targetId) ||
                other.targetId == targetId));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, questionId, targetId);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VoteInfoImplCopyWith<_$VoteInfoImpl> get copyWith =>
      __$$VoteInfoImplCopyWithImpl<_$VoteInfoImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VoteInfoImplToJson(
      this,
    );
  }
}

abstract class _VoteInfo implements VoteInfo {
  const factory _VoteInfo(
          {@JsonKey(name: 'question_id') required final String questionId,
          @JsonKey(name: 'target_id') required final String targetId}) =
      _$VoteInfoImpl;

  factory _VoteInfo.fromJson(Map<String, dynamic> json) =
      _$VoteInfoImpl.fromJson;

  @override
  @JsonKey(name: 'question_id')
  String get questionId;
  @override
  @JsonKey(name: 'target_id')
  String get targetId;
  @override
  @JsonKey(ignore: true)
  _$$VoteInfoImplCopyWith<_$VoteInfoImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

GroupVotingProgress _$GroupVotingProgressFromJson(Map<String, dynamic> json) {
  return _GroupVotingProgress.fromJson(json);
}

/// @nodoc
mixin _$GroupVotingProgress {
  @JsonKey(name: 'voted_count')
  int get votedCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'total_count')
  int get totalCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'quorum_reached')
  bool get quorumReached => throw _privateConstructorUsedError;
  @JsonKey(name: 'quorum_threshold')
  double get quorumThreshold => throw _privateConstructorUsedError;
  @JsonKey(name: 'user_voted')
  bool get userVoted => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $GroupVotingProgressCopyWith<GroupVotingProgress> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $GroupVotingProgressCopyWith<$Res> {
  factory $GroupVotingProgressCopyWith(
          GroupVotingProgress value, $Res Function(GroupVotingProgress) then) =
      _$GroupVotingProgressCopyWithImpl<$Res, GroupVotingProgress>;
  @useResult
  $Res call(
      {@JsonKey(name: 'voted_count') int votedCount,
      @JsonKey(name: 'total_count') int totalCount,
      @JsonKey(name: 'quorum_reached') bool quorumReached,
      @JsonKey(name: 'quorum_threshold') double quorumThreshold,
      @JsonKey(name: 'user_voted') bool userVoted});
}

/// @nodoc
class _$GroupVotingProgressCopyWithImpl<$Res, $Val extends GroupVotingProgress>
    implements $GroupVotingProgressCopyWith<$Res> {
  _$GroupVotingProgressCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? votedCount = null,
    Object? totalCount = null,
    Object? quorumReached = null,
    Object? quorumThreshold = null,
    Object? userVoted = null,
  }) {
    return _then(_value.copyWith(
      votedCount: null == votedCount
          ? _value.votedCount
          : votedCount // ignore: cast_nullable_to_non_nullable
              as int,
      totalCount: null == totalCount
          ? _value.totalCount
          : totalCount // ignore: cast_nullable_to_non_nullable
              as int,
      quorumReached: null == quorumReached
          ? _value.quorumReached
          : quorumReached // ignore: cast_nullable_to_non_nullable
              as bool,
      quorumThreshold: null == quorumThreshold
          ? _value.quorumThreshold
          : quorumThreshold // ignore: cast_nullable_to_non_nullable
              as double,
      userVoted: null == userVoted
          ? _value.userVoted
          : userVoted // ignore: cast_nullable_to_non_nullable
              as bool,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$GroupVotingProgressImplCopyWith<$Res>
    implements $GroupVotingProgressCopyWith<$Res> {
  factory _$$GroupVotingProgressImplCopyWith(_$GroupVotingProgressImpl value,
          $Res Function(_$GroupVotingProgressImpl) then) =
      __$$GroupVotingProgressImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'voted_count') int votedCount,
      @JsonKey(name: 'total_count') int totalCount,
      @JsonKey(name: 'quorum_reached') bool quorumReached,
      @JsonKey(name: 'quorum_threshold') double quorumThreshold,
      @JsonKey(name: 'user_voted') bool userVoted});
}

/// @nodoc
class __$$GroupVotingProgressImplCopyWithImpl<$Res>
    extends _$GroupVotingProgressCopyWithImpl<$Res, _$GroupVotingProgressImpl>
    implements _$$GroupVotingProgressImplCopyWith<$Res> {
  __$$GroupVotingProgressImplCopyWithImpl(_$GroupVotingProgressImpl _value,
      $Res Function(_$GroupVotingProgressImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? votedCount = null,
    Object? totalCount = null,
    Object? quorumReached = null,
    Object? quorumThreshold = null,
    Object? userVoted = null,
  }) {
    return _then(_$GroupVotingProgressImpl(
      votedCount: null == votedCount
          ? _value.votedCount
          : votedCount // ignore: cast_nullable_to_non_nullable
              as int,
      totalCount: null == totalCount
          ? _value.totalCount
          : totalCount // ignore: cast_nullable_to_non_nullable
              as int,
      quorumReached: null == quorumReached
          ? _value.quorumReached
          : quorumReached // ignore: cast_nullable_to_non_nullable
              as bool,
      quorumThreshold: null == quorumThreshold
          ? _value.quorumThreshold
          : quorumThreshold // ignore: cast_nullable_to_non_nullable
              as double,
      userVoted: null == userVoted
          ? _value.userVoted
          : userVoted // ignore: cast_nullable_to_non_nullable
              as bool,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$GroupVotingProgressImpl implements _GroupVotingProgress {
  const _$GroupVotingProgressImpl(
      {@JsonKey(name: 'voted_count') required this.votedCount,
      @JsonKey(name: 'total_count') required this.totalCount,
      @JsonKey(name: 'quorum_reached') required this.quorumReached,
      @JsonKey(name: 'quorum_threshold') required this.quorumThreshold,
      @JsonKey(name: 'user_voted') required this.userVoted});

  factory _$GroupVotingProgressImpl.fromJson(Map<String, dynamic> json) =>
      _$$GroupVotingProgressImplFromJson(json);

  @override
  @JsonKey(name: 'voted_count')
  final int votedCount;
  @override
  @JsonKey(name: 'total_count')
  final int totalCount;
  @override
  @JsonKey(name: 'quorum_reached')
  final bool quorumReached;
  @override
  @JsonKey(name: 'quorum_threshold')
  final double quorumThreshold;
  @override
  @JsonKey(name: 'user_voted')
  final bool userVoted;

  @override
  String toString() {
    return 'GroupVotingProgress(votedCount: $votedCount, totalCount: $totalCount, quorumReached: $quorumReached, quorumThreshold: $quorumThreshold, userVoted: $userVoted)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$GroupVotingProgressImpl &&
            (identical(other.votedCount, votedCount) ||
                other.votedCount == votedCount) &&
            (identical(other.totalCount, totalCount) ||
                other.totalCount == totalCount) &&
            (identical(other.quorumReached, quorumReached) ||
                other.quorumReached == quorumReached) &&
            (identical(other.quorumThreshold, quorumThreshold) ||
                other.quorumThreshold == quorumThreshold) &&
            (identical(other.userVoted, userVoted) ||
                other.userVoted == userVoted));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, votedCount, totalCount,
      quorumReached, quorumThreshold, userVoted);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$GroupVotingProgressImplCopyWith<_$GroupVotingProgressImpl> get copyWith =>
      __$$GroupVotingProgressImplCopyWithImpl<_$GroupVotingProgressImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$GroupVotingProgressImplToJson(
      this,
    );
  }
}

abstract class _GroupVotingProgress implements GroupVotingProgress {
  const factory _GroupVotingProgress(
      {@JsonKey(name: 'voted_count') required final int votedCount,
      @JsonKey(name: 'total_count') required final int totalCount,
      @JsonKey(name: 'quorum_reached') required final bool quorumReached,
      @JsonKey(name: 'quorum_threshold') required final double quorumThreshold,
      @JsonKey(name: 'user_voted')
      required final bool userVoted}) = _$GroupVotingProgressImpl;

  factory _GroupVotingProgress.fromJson(Map<String, dynamic> json) =
      _$GroupVotingProgressImpl.fromJson;

  @override
  @JsonKey(name: 'voted_count')
  int get votedCount;
  @override
  @JsonKey(name: 'total_count')
  int get totalCount;
  @override
  @JsonKey(name: 'quorum_reached')
  bool get quorumReached;
  @override
  @JsonKey(name: 'quorum_threshold')
  double get quorumThreshold;
  @override
  @JsonKey(name: 'user_voted')
  bool get userVoted;
  @override
  @JsonKey(ignore: true)
  _$$GroupVotingProgressImplCopyWith<_$GroupVotingProgressImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
